package mongosearch

import (
	"github.com/300brand/logger"
	"github.com/300brand/searchquery"
	"labix.org/v2/mgo/bson"
)

func (s *MongoSearch) buildQuery(query *searchquery.Query) (mgoQuery bson.M, err error) {

	return
}

/*
func (s *MongoSearch) buildQuery(query *searchquery.Query) (mgoQuery bson.M, err error) {
	// logger.Trace.Printf("buildQuery: Req:%d Opt:%d Exc:%d", len(query.Required), len(query.Optional), len(query.Excluded))
	mgoQuery = bson.M{}

	if err = s.loopSubqueries(query.Required, "$and", mgoQuery); err != nil {
		return
	}
	if err = s.loopSubqueries(query.Optional, "$or", mgoQuery); err != nil {
		return
	}
	// if err = s.loopSubqueries(query.Excluded, "$nor", mgoQuery); err != nil {
	// 	return
	// }
	if len(query.Excluded) > 0 {
		s.reqMapReduce = true
	}
	return
}

func (s *MongoSearch) buildSubquery(subquery *searchquery.SubQuery) (mgoSubquery bson.M, err error) {
	// logger.Trace.Printf("buildSubquery: %s %s %s", subquery.Field, subquery.Operator, subquery.Value)

	if subquery.Query != nil {
		return s.buildQuery(subquery.Query)
	}

	errInvalidOp := "Cannot use %s operator with an array value for %s"

	field, value, isArray, err := s.realValue(subquery)
	if err != nil {
		return
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

func (s *MongoSearch) canOptimize(subqueries []searchquery.SubQuery) bool {
	if len(subqueries) == 0 {
		return false
	}

	field, _, _, _ := s.realValue(&subqueries[0])
	for _, sq := range subqueries {
		if sq.Query != nil {
			// logger.Trace.Printf("canOptimize: sq.Query != nil")
			return false
		}

		var err error
		var isArray bool
		sqField := sq.Field

		if newName, ok := s.Rewrites[sqField]; ok {
			sqField = newName
		}

		if field != sqField {
			// logger.Trace.Printf("canOptimize: %s != %s", field, sqField)
			return false
		}

		if convertFunc, ok := s.Fields[sqField]; ok {
			if _, isArray, err = convertFunc(sq.Value); err != nil {
				// logger.Trace.Printf("canOptimize: %s returned error - %s", sqField, err)
				return false
			}
		}

		if isArray {
			// logger.Trace.Printf("canOptimize: %s is array", sq)
			s.reqMapReduce = true
			return false
		}
	}

	return true
}

func (s *MongoSearch) loopSubqueries(subqueries []searchquery.SubQuery, op string, into bson.M) (err error) {
	if len(subqueries) == 0 {
		return
	}

	if s.canOptimize(subqueries) {
		var field string
		// logger.Trace.Printf("loopSubqueries: canOptimize")
		if len(subqueries) == 1 {
			var value interface{}
			field, value, _, err = s.realValue(&subqueries[0])
			into[field] = value
			return
		}

		values := make([]interface{}, len(subqueries))
		for i := range subqueries {
			field, values[i], _, _ = s.realValue(&subqueries[i])
		}

		switch op {
		case "$or":
			into[field] = bson.M{"$in": values}
		case "$and":
			into[field] = bson.M{"$all": values}
		}
		return
	}

	// logger.Trace.Printf("loopSubqueries: Making subs for %s with len: %d", op, len(subqueries))
	subs := make([]bson.M, 0, len(subqueries))
	for _, sq := range subqueries {
		built, err := s.buildSubquery(&sq)
		if err != nil {
			return err
		}
		subs = append(subs, built)
	}
	into[op] = subs
	return
}

func (s *MongoSearch) realValue(subquery *searchquery.SubQuery) (field string, value interface{}, isArray bool, err error) {
	field = subquery.Field
	value = subquery.Value

	if newName, ok := s.Rewrites[field]; ok {
		field = newName
	}

	if convertFunc, ok := s.Fields[field]; ok {
		if value, isArray, err = convertFunc(subquery.Value); err != nil {
			err = fmt.Errorf("Error converting %s: %s", field, err)
			return
		}
	}

	return
}
*/
