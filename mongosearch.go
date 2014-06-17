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
	CollItems     string // Collection of items to search
	CollResults   string // Search resutls collection
	FieldAllWords string // Items: Path to array of individual words
	FieldDate     string // Items: Path to item date
	FieldKeywords string // Items: Path to keywords array
	UrlItems      string // Connection string to Items database and collection: host:port/db
	UrlResults    string // Connection string to Results database and collection: host:port/db
	cItems        *mgo.Collection
	cResults      *mgo.Collection
}

var StopWords = strings.Fields(`
	a about above after again against all am an and any are arent as at be
	because been before being below between both but by cant cannot could
	couldnt did didnt do does doesnt doing dont down during each few for
	from further had hadnt has hasnt have havent having he hed hell hes her
	here heres hers herself him himself his how hows i id ill im ive if in
	into is isnt it its its itself lets me more most mustnt my myself no nor
	not of off on once only or other ought our ours ourselves out over own
	same shant she shed shell shes should shouldnt so some such than that
	thats the their theirs them themselves then there theres these they
	theyd theyll theyre theyve this those through to too under until up very
	was wasnt we wed well were weve were werent what whats when whens where
	wheres which while who whos whom why whys with wont would wouldnt you
	youd youll youre youve your yours yourself yourselves
`)
var TimeLayout = "2006-01-02 15:04:05"

func New(urlItems, cItems, urlResults, cResults, fAllWords, fDate, fKeywords string) (s *MongoSearch, err error) {
	s = &MongoSearch{
		CollItems:     cItems,
		CollResults:   cResults,
		FieldAllWords: fAllWords,
		FieldDate:     fDate,
		FieldKeywords: fKeywords,
		UrlItems:      urlItems,
		UrlResults:    urlResults,
	}
	sess, err := mgo.Dial(urlItems)
	if err != nil {
		return
	}
	s.cItems = sess.DB("").C(cItems)

	sess, err = mgo.Dial(urlResults)
	if err != nil {
		return
	}
	s.cResults = sess.DB("").C(cResults)
	return
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

func (s *MongoSearch) doSearch(query string, filter bson.M, id bson.ObjectId) (err error) {
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
	if _, err = s.cResults.UpsertId(id, bson.M{
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

	return
}

func (s *MongoSearch) buildQuery(query *searchquery.Query) (mgoQuery bson.M, err error) {
	logger.Trace.Printf("buildQuery: R:%d O:%d E:%d", len(query.Required), len(query.Optional), len(query.Excluded))
	mgoQuery = bson.M{}
	loop := func(subQueries []searchquery.SubQuery, op string) (err error) {
		if len(subQueries) > 0 {
			logger.Trace.Printf("Making subs for %s with len: %d", op, len(subQueries))
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
	if err = loop(query.Excluded, "$not"); err != nil {
		return
	}
	return
}

func (s *MongoSearch) buildSubquery(subquery *searchquery.SubQuery) (mgoSubquery bson.M, err error) {
	logger.Trace.Printf("buildSubquery: %s %s %s", subquery.Field, subquery.Operator, subquery.Value)
	if subquery.Query != nil {
		return s.buildQuery(subquery.Query)
	}

	var value interface{} = subquery.Value
	field := subquery.Field
	switch field {
	case s.FieldDate:
		value, err = time.Parse(TimeLayout, subquery.Value)
		if err != nil {
			return nil, err
		}
	case "":
		field = s.FieldKeywords
	default:
		err = fmt.Errorf("Unknown field: %s", subquery.Field)
		return
	}
	switch subquery.Operator {
	case searchquery.OperatorRelE, searchquery.OperatorField:
		if field == s.FieldKeywords {
			value = bson.M{"$all": strings.Fields(subquery.Value)}
		}
	case searchquery.OperatorRelGT:
		value = bson.M{"$gt": value}
	case searchquery.OperatorRelGTE:
		value = bson.M{"$gte": value}
	case searchquery.OperatorRelLT:
		value = bson.M{"$lt": value}
	case searchquery.OperatorRelLTE:
		value = bson.M{"$lte": value}
	case searchquery.OperatorRelNE:
		value = bson.M{"$ne": value}
	default:
		err = fmt.Errorf("Unknown operator: %s", subquery.Operator)
		return
	}
	mgoSubquery = bson.M{
		field: value,
	}
	return
}
