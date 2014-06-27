package mongosearch

import (
	"encoding/json"
	"fmt"
	"github.com/300brand/logger"
	"github.com/300brand/searchquery"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

type MongoSearch struct {
	CollItems    string                    // Collection of items to search
	CollResults  string                    // Search resutls collection
	Conversions  map[string]ConversionFunc // Field -> ConversionFunc map; if field not found, entire string used
	Rewrites     map[string]string         // Rewrite rules for final query output (allows simpler inbound queries and rewrite of default "" field)
	Url          string                    // Connection string to database: host:port/db
	reqMapReduce bool
	fields       struct {
		all     string
		keyword string
		pubdate string
		pubid   string
	}
}

var TimeLayout = "2006-01-02"

// serverUrl - Yup.
// cItems    -
// cResults  - Collection name in the form <db>.<coll> or <coll> (db from
//             connection string is uesd)
func New(serverUrl, cItems, cResults string) (s *MongoSearch, err error) {
	s = &MongoSearch{
		CollItems:   cItems,
		CollResults: cResults,
		Url:         serverUrl,
	}
	s.Conversions = make(map[string]ConversionFunc)
	s.Rewrites = make(map[string]string)
	return
}

func (s *MongoSearch) SetAll(name string) {
	s.fields.all = name
}

func (s *MongoSearch) SetKeyword(name string, convertFunc ConversionFunc, aliases ...string) {
	for _, alias := range aliases {
		s.Rewrite(alias, name)
	}
	s.Rewrite(name, name)
	s.Convert(name, convertFunc)
	s.fields.keyword = name
}

func (s *MongoSearch) SetPubdate(name string, convertFunc ConversionFunc, aliases ...string) {
	for _, alias := range aliases {
		s.Rewrite(alias, name)
	}
	s.Rewrite(name, name)
	s.Convert(name, convertFunc)
	s.fields.pubdate = name
}

func (s *MongoSearch) SetPubid(name string, convertFunc ConversionFunc, aliases ...string) {
	for _, alias := range aliases {
		s.Rewrite(alias, name)
	}
	s.Rewrite(name, name)
	s.Convert(name, convertFunc)
	s.fields.pubid = name
}

func (s *MongoSearch) Convert(field string, convertFunc ConversionFunc) {
	s.Conversions[field] = convertFunc
}

func (s *MongoSearch) Rewrite(field, newName string) {
	s.Rewrites[field] = newName
}

func (s *MongoSearch) Search(query string) (id bson.ObjectId, err error) {
	id = bson.NewObjectId()
	err = s.doSearch(query, id)
	return
}

func (s *MongoSearch) SearchInto(query string, id bson.ObjectId) (err error) {
	return s.doSearch(query, id)
}

func (s *MongoSearch) dbFor(session *mgo.Session, collection string) (db, coll string) {
	bits := strings.SplitN(collection, ".", 2)
	if len(bits) == 1 {
		return session.DB("").Name, bits[0]
	}
	return bits[0], bits[1]
}

func (s *MongoSearch) buildScope(query *searchquery.Query) (scope bson.M, err error) {
	// logger.Trace.Printf("buildScope: R:%d O:%d E:%d", len(query.Required), len(query.Optional), len(query.Excluded))
	scope = bson.M{}
	loop := func(subQueries []searchquery.SubQuery, op string) (err error) {
		if len(subQueries) > 0 {
			// logger.Trace.Printf("buildScope: Making subs for %s with len: %d", op, len(subQueries))
			subs := make([]interface{}, 0, len(subQueries))
			for _, sq := range subQueries {
				built, err := s.buildSubscope(&sq)
				if err != nil {
					return err
				}
				if built == nil {
					continue
				}
				subs = append(subs, built)
			}
			scope[op] = subs
		}
		return
	}
	if err = loop(query.Required, "and"); err != nil {
		return
	}
	if err = loop(query.Optional, "or"); err != nil {
		return
	}
	if err = loop(query.Excluded, "nor"); err != nil {
		return
	}
	return
}

func (s *MongoSearch) buildSubscope(subquery *searchquery.SubQuery) (subscope interface{}, err error) {
	// logger.Trace.Printf("buildSubscope: %s %s %s", subquery.Field, subquery.Operator, subquery.Value)

	if subquery.Query != nil {
		return s.buildScope(subquery.Query)
	}

	if name, ok := s.Rewrites[subquery.Field]; !ok || name != s.fields.keyword {
		return
	}

	return subquery.Value, nil
}

func (s *MongoSearch) doMapReduce(session *mgo.Session, query *searchquery.Query, id bson.ObjectId) (info *mgo.MapReduceInfo, err error) {
	// logger.Trace.Printf("doMapReduce: starting")
	mgoQuery, err := s.buildQuery(query)
	if err != nil {
		return
	}
	// logger.Trace.Printf("doMapReduce: mgoQuery: %+v", mgoQuery)
	scope, err := s.buildScope(query)
	if err != nil {
		return
	}
	// logger.Trace.Printf("doMapReduce: scope: %+v", scope)

	db, coll := s.dbFor(session, s.CollResults)
	coll = fmt.Sprintf("%s_%s", coll, id.Hex())

	job := &mgo.MapReduce{
		Reduce: `function(key, values) { return values[0] }`,
		Out: bson.M{
			"replace": coll,
			"db":      db,
		},
		Scope: bson.M{
			"query": scope,
		},
		Verbose: true,
	}

	if s.reqMapReduce {
		job.Map = fmt.Sprintf(mapFunc, s.fields.all)
	} else {
		job.Map = mapFuncImmediate
	}

	db, coll = s.dbFor(session, s.CollItems)
	return session.DB(db).C(coll).Find(mgoQuery).MapReduce(job, nil)
}

func (s *MongoSearch) doSearch(query string, id bson.ObjectId) (err error) {
	// Check if all the fields are defined
	switch "" {
	case s.fields.all:
		return fmt.Errorf("Use SetAll() to define a value for the all-words array")
	case s.fields.keyword:
		return fmt.Errorf("Use SetKeyword() to define a value for the all-words array")
	case s.fields.pubdate:
		return fmt.Errorf("Use SetPubdate() to define a value for the all-words array")
	case s.fields.pubid:
		return fmt.Errorf("Use SetPubid() to define a value for the all-words array")
	}

	session, err := mgo.Dial(s.Url)
	if err != nil {
		return
	}
	defer session.Close()

	session.SetSocketTimeout(60 * time.Minute)

	db, coll := s.dbFor(session, s.CollResults)

	q, err := searchquery.ParseGreedy(query)
	if err != nil {
		return
	}

	// logger.Debug.Printf("Query: %+v", q)
	built, err := s.buildQuery(q)
	if err != nil {
		return
	}
	jsonBuilt, _ := json.Marshal(built)
	logger.Info.Printf("Parsed: %s", jsonBuilt)

	if _, err = session.DB(db).C(coll).UpsertId(id, bson.M{
		"$set": bson.M{
			"query": bson.M{
				"original": query,
				"parsed":   q.String(),
			},
			"doMapReduce": s.reqMapReduce,
			"start":       time.Now(),
		},
	}); err != nil {
		return
	}

	info, err := s.doMapReduce(session, q, id)
	if err != nil {
		return
	}

	if err = session.DB(db).C(coll).UpdateId(id, bson.M{
		"$set": bson.M{
			"end":  time.Now(),
			"info": info,
		},
	}); err != nil {
		return
	}

	return
}
