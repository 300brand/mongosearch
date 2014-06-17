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
	UrlItems    string                // Connection string to Items database and collection: host:port/db
	UrlResults  string                // Connection string to Results database and collection: host:port/db
	cItems      *mgo.Collection
	cResults    *mgo.Collection
}

var TimeLayout = "2006-01-02 15:04:05"

func New(urlItems, cItems, urlResults, cResults string) (s *MongoSearch, err error) {
	s = &MongoSearch{
		CollItems:   cItems,
		CollResults: cResults,
		UrlItems:    urlItems,
		UrlResults:  urlResults,
	}
	s.Fields = map[string]Conversion{
		"": ConvertStops,
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

func (s *MongoSearch) Convert(field string, convertFunc Conversion) {
	s.Fields[field] = convertFunc
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

	var (
		isArray      bool
		errInvalidOp             = "Cannot use %s operator with an array value for %s"
		field                    = subquery.Field
		value        interface{} = subquery.Value
	)

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
