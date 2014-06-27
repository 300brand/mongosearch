package mongosearch

import (
	"fmt"
	"github.com/300brand/logger"
	"github.com/300brand/searchquery"
	"labix.org/v2/mgo/bson"
	"time"
)

func (s *MongoSearch) buildQuery(query *searchquery.Query) (mgoQuery bson.M, err error) {
	// logger.Info.Printf("buildQuery: starting with %s", query)

	fields := s.mapFields(query)

	keywordSubquery, ok := fields[s.fields.keyword]
	if !ok {
		err = fmt.Errorf("No field found for %s", s.fields.keyword)
		return
	}
	dateSubquery, ok := fields[s.fields.pubdate]
	if !ok {
		dateSubquery = searchquery.SubQuery{
			Value:    time.Now().Format(TimeLayout),
			Field:    s.fields.pubdate,
			Operator: searchquery.OperatorRelE,
		}
	}

	reduced := s.reduce(keywordSubquery)

	subqueries := make([]searchquery.SubQuery, 0, len(reduced.Optional)+len(reduced.Required))
	subqueries = append(subqueries, reduced.Optional...)
	subqueries = append(subqueries, reduced.Required...)

	// Remove the keyword field as it gets split up for stacking. Remaining
	// fields are added to each keyword component
	delete(fields, s.fields.keyword)
	delete(fields, s.fields.pubdate)

	// Convert values
	convertedFields := make(map[string]bson.M, len(fields))
	for fName, f := range fields {
		convertedFields[fName], err = s.convertSubquery(&f)
		if err != nil {
			logger.Error.Printf("buildQuery: %s", err)
			return
		}
	}

	dateValue, err := s.convertSubquery(&dateSubquery)
	if err != nil {
		return
	}

	// This is a pain in the ass..
	var dates []interface{}

	dateIn, ok := dateValue[s.fields.pubdate]
	if !ok {
		return nil, fmt.Errorf("dateValue has no field %s: %#v", s.fields.pubdate, dateValue)
	}

	switch t := dateIn.(type) {
	case bson.M:
		if dates, ok = t["$in"].([]interface{}); !ok {
			return nil, fmt.Errorf("Crazy setup in the dateIn struct %#v", dateIn)
		}
	case interface{}:
		dates = []interface{}{t}
	default:
		return nil, fmt.Errorf("Weird problem with dateIn arr... %#v", dateIn)
	}

	mgoSubs := make([]bson.M, len(subqueries)*len(dates))
	for i, date := range dates {
		for j, subquery := range subqueries {
			idx := i*len(subqueries) + j
			mgoSubs[idx] = make(bson.M, len(convertedFields)+1)

			// Set date
			mgoSubs[idx][s.fields.pubdate] = date

			// logger.Warn.Printf("%#v", subquery)
			// Process the keywords
			if subquery.Operator == searchquery.OperatorSubquery {
				value, err := s.convertQuery(s.reduce(subquery))
				if err != nil {
					return nil, err
				}
				for k := range value {
					mgoSubs[idx][k] = value[k]
				}
				// logger.Error.Printf("convertQuery: %#v", value)
			} else {
				value, err := s.convertSubquery(&subquery)
				if err != nil {
					return nil, err
				}
				// logger.Warn.Printf("%#v", value)
				for k := range value {
					mgoSubs[idx][k] = value[k]
				}
			}

			// Push in remaining fields
			// TODO cleanup
			for k, v := range convertedFields {
				mgoSubs[idx][k] = v[k]
			}
		}
	}
	mgoQuery = bson.M{"$or": mgoSubs}

	return
}

func (s *MongoSearch) mapFields(query *searchquery.Query) (fields map[string]searchquery.SubQuery) {
	fields = make(map[string]searchquery.SubQuery)

	querySubs := [][]searchquery.SubQuery{
		query.Required,
		query.Optional,
	}

	for _, subqueries := range querySubs {
		for i := range subqueries {
			// Blank field names denote an outer set of parenthesis; recurse
			// into these to restart the field search
			name := subqueries[i].Field
			if name != "" {
				if newName, ok := s.Rewrites[name]; ok {
					name = newName
				}
				if _, ok := fields[name]; !ok {
					fields[name] = subqueries[i]
					continue
				}
			}
			// After recursing into each subquery's query; merge results into
			// field map
			if q := subqueries[i].Query; q != nil {
				merge := s.mapFields(q)
				for k := range merge {
					if _, ok := fields[k]; !ok {
						fields[k] = merge[k]
					}
				}
			}
		}
	}

	return
}

func (s *MongoSearch) reduce(subquery searchquery.SubQuery) (reduced *searchquery.Query) {
	for subquery.Operator == searchquery.OperatorSubquery {
		reduced = subquery.Query
		if len(reduced.Excluded) > 0 {
			s.reqMapReduce = true
		}
		if len(reduced.Optional)+len(reduced.Required) > 1 {
			break
		}

		if subs := reduced.Optional; len(subs) == 1 {
			subquery = subs[0]
		}
		if subs := reduced.Required; len(subs) == 1 {
			subquery = subs[0]
		}
	}

	if reduced == nil {
		reduced = &searchquery.Query{
			Optional: []searchquery.SubQuery{subquery},
		}
	}

	optreqs := [][]searchquery.SubQuery{reduced.Optional, reduced.Required}
	for _, optreq := range optreqs {
		if optreq == nil {
			continue
		}
		for i := range optreq {
			for optreq[i].Operator == searchquery.OperatorSubquery {
				if subs := optreq[i].Query.Optional; len(subs) == 1 {
					optreq[i] = subs[0]
				} else if len(subs) > 1 {
					break
				}
				if subs := optreq[i].Query.Required; len(subs) == 1 {
					optreq[i] = subs[0]
				} else if len(subs) > 1 {
					break
				}
			}
		}
	}

	return
}

func (s *MongoSearch) convertQuery(query *searchquery.Query) (mgoQuery bson.M, err error) {
	// logger.Trace.Printf("convertQuery: Req:%d Opt:%d Exc:%d", len(query.Required), len(query.Optional), len(query.Excluded))
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
	return
}

func (s *MongoSearch) convertSubquery(subquery *searchquery.SubQuery) (mgoSubquery bson.M, err error) {
	// logger.Trace.Printf("buildSubquery: %s %s %s", subquery.Field, subquery.Operator, subquery.Value)

	if subquery.Query != nil {
		return s.convertQuery(subquery.Query)
	}

	errInvalidOp := "Cannot use %s operator with an array value for %s"

	field, value, isArray, err := s.realValue(subquery)
	if err != nil {
		return
	}
	if isArray {
		s.reqMapReduce = true
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

		if convertFunc, ok := s.Conversions[sqField]; ok {
			if _, isArray, err = convertFunc(sq.Value); err != nil {
				// logger.Trace.Printf("canOptimize: %s returned error - %s", sqField, err)
				return false
			}
		}

		if isArray {
			// logger.Trace.Printf("canOptimize: %s is array", sq)
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
		built, err := s.convertSubquery(&sq)
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

	if convertFunc, ok := s.Conversions[field]; ok {
		if value, isArray, err = convertFunc(subquery.Value); err != nil {
			err = fmt.Errorf("Error converting %s: %s", field, err)
			return
		}
	}

	return
}
