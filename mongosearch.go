package mongosearch

import (
	"fmt"
	"github.com/300brand/logger"
	"github.com/300brand/searchquery"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

type MongoSearch struct {
	CollItems   string                // Collection of items to search
	CollResults string                // Search resutls collection
	Fields      map[string]Conversion // Field -> Conversion map; if field not found, entire string used
	Rewrites    map[string]string     // Rewrite rules for final query output (allows simpler inbound queries and rewrite of default "" field)
	Url         string                // Connection string to database: host:port/db
}

var TimeLayout = "2006-01-02 15:04:05"

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
	s.Fields = map[string]Conversion{
		"": ConvertSpaces,
	}
	s.Rewrites = make(map[string]string)
	return
}

func (s *MongoSearch) Convert(field string, convertFunc Conversion) {
	s.Fields[field] = convertFunc
}

func (s *MongoSearch) Rewrite(field, newName string) {
	s.Rewrites[field] = newName
}

func (s *MongoSearch) Search(query string) (id bson.ObjectId, err error) {
	id = bson.NewObjectId()
	err = s.doSearch(query, bson.M{}, id)
	return
}

func (s *MongoSearch) SearchFilter(query string, filter bson.M) (id bson.ObjectId, err error) {
	if filter == nil {
		filter = bson.M{}
	}
	id = bson.NewObjectId()
	err = s.doSearch(query, filter, id)
	return
}

func (s *MongoSearch) SearchFilterInto(query string, filter bson.M, id bson.ObjectId) (err error) {
	if filter == nil {
		filter = bson.M{}
	}
	id = bson.NewObjectId()
	err = s.doSearch(query, filter, id)
	return
}

func (s *MongoSearch) SearchInto(query string, id bson.ObjectId) (err error) {
	return s.doSearch(query, bson.M{}, id)
}

func (s *MongoSearch) dbFor(session *mgo.Session, collection string) (db, coll string) {
	bits := strings.SplitN(collection, ".", 2)
	if len(bits) == 1 {
		return session.DB("").Name, bits[0]
	}
	return bits[0], bits[1]
}

func (s *MongoSearch) doSearch(query string, filter bson.M, id bson.ObjectId) (err error) {
	sess, err := mgo.Dial(s.Url)
	if err != nil {
		return
	}
	defer sess.Close()

	db, coll := s.dbFor(sess, s.CollResults)

	q, err := searchquery.Parse(query)
	if err != nil {
		return
	}
	logger.Info.Printf("Query: %+v", q)
	a, err := s.buildQuery(q)
	if err != nil {
		return
	}
	logger.Info.Printf("Aggregate: %+v", a)
	if _, err = sess.DB(db).C(coll).UpsertId(id, bson.M{
		"$set": bson.M{
			"query": bson.M{
				"original": query,
				"parsed":   q.String(),
			},
			"start": time.Now(),
		},
	}); err != nil {
		return
	}

	if _, err = sess.DB(db).C(coll).UpdateId(id, bson.M{
		"$set": bson.M{
			"end":  time.Now(),
			"info": info,
		},
	}); err != nil {
		return
	}

	return
}

func (s *MongoSearch) buildQuery(query *searchquery.Query) (mgoQuery bson.M, err error) {
	// logger.Trace.Printf("buildQuery: R:%d O:%d E:%d", len(query.Required), len(query.Optional), len(query.Excluded))
	mgoQuery = bson.M{}
	loop := func(subQueries []searchquery.SubQuery, op string) (err error) {
		if len(subQueries) > 0 {
			// logger.Trace.Printf("buildQuery: Making subs for %s with len: %d", op, len(subQueries))
			subs := make([]bson.M, 0, len(subQueries))
			for _, sq := range subQueries {
				built, err := s.buildSubquery(&sq)
				if err != nil {
					return err
				}
				subs = append(subs, built)
			}
			mgoQuery[op] = subs
		}
		return
	}
	if err = loop(query.Required, "$and"); err != nil {
		return
	}
	if err = loop(query.Optional, "$or"); err != nil {
		return
	}
	// if err = loop(query.Excluded, "$nor"); err != nil {
	// 	return
	// }
	return
}

func (s *MongoSearch) buildSubquery(subquery *searchquery.SubQuery) (mgoSubquery bson.M, err error) {
	// logger.Trace.Printf("buildSubquery: %s %s %s", subquery.Field, subquery.Operator, subquery.Value)

	if subquery.Query != nil {
		return s.buildQuery(subquery.Query)
	}

	var (
		isArray      bool
		errInvalidOp             = "Cannot use %s operator with an array value for %s"
		field                    = subquery.Field
		value        interface{} = subquery.Value
	)

	if newName, ok := s.Rewrites[field]; ok {
		field = newName
	}

	if convertFunc, ok := s.Fields[field]; ok {
		if value, isArray, err = convertFunc(subquery.Value); err != nil {
			err = fmt.Errorf("Error converting %s: %s", field, err)
			return
		}
	}

	// Wrap value in proper operator
	switch subquery.Operator {
	case searchquery.OperatorRelE, searchquery.OperatorField:
		if isArray {
			value = bson.M{"$all": value}
		}
		// value = value for scalar
	case searchquery.OperatorRelGT:
		if isArray {
			err = fmt.Errorf(errInvalidOp, subquery.Operator, field)
			return
		}
		value = bson.M{"$gt": value}
	case searchquery.OperatorRelGTE:
		if isArray {
			err = fmt.Errorf(errInvalidOp, subquery.Operator, field)
			return
		}
		value = bson.M{"$gte": value}
	case searchquery.OperatorRelLT:
		if isArray {
			err = fmt.Errorf(errInvalidOp, subquery.Operator, field)
			return
		}
		value = bson.M{"$lt": value}
	case searchquery.OperatorRelLTE:
		if isArray {
			err = fmt.Errorf(errInvalidOp, subquery.Operator, field)
			return
		}
		value = bson.M{"$lte": value}
	case searchquery.OperatorRelNE:
		if isArray {
			value = bson.M{"$nin": value}
		} else {
			value = bson.M{"$ne": value}
		}
	default:
		err = fmt.Errorf("Unknown operator: %s", subquery.Operator)
		return
	}
	mgoSubquery = bson.M{
		field: value,
	}
	return
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

	if subquery.Field != "" {
		return
	}

	return subquery.Value, nil
}

func (s *MongoSearch) doMapReduce(session *mgo.Session, query *searchquery.Query, id bson.ObjectId) (info *mgo.MapReduceInfo, err error) {
	mgoQuery, err := s.buildQuery(query)
	if err != nil {
		return
	}
	scope, err := s.buildScope(query)
	if err != nil {
		return
	}

	db, coll := s.dbFor(session, s.CollResults)
	coll = fmt.Sprintf("%s_%s", coll, id.Hex())

	job := &mgo.MapReduce{
		Map:    mapFunc,
		Reduce: `function(key, values) { return values[0] }`,
		Out: bson.M{
			"replace": coll,
			"db":      db,
		},
		Scope:   scope,
		Verbose: true,
	}

	db, coll = s.dbFor(session, s.CollItems)
	return session.DB(db).C(coll).Find(mgoQuery).MapReduce(job, nil)
}
