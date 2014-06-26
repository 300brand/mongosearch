package mongosearch

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

type ConversionFunc func(string) (interface{}, bool, error)

var ConvertDate ConversionFunc = func(in string) (out interface{}, isArray bool, err error) {
	out, err = time.Parse(TimeLayout, in)
	return
}

var ConvertBsonId ConversionFunc = func(in string) (out interface{}, isArray bool, err error) {
	if !bson.IsObjectIdHex(in) {
		err = fmt.Errorf("Invalid BSON ObjectId: %s", in)
		return
	}
	out, err = bson.ObjectIdHex(in), nil
	return
}

var ConvertSpaces ConversionFunc = func(in string) (out interface{}, isArray bool, err error) {
	if in == "" {
		return
	}

	fields := strings.Fields(in)
	if isArray = len(fields) > 1; isArray {
		out = fields
	} else {
		out = fields[0]
	}
	return
}

var ConvertDateInt ConversionFunc = func(in string) (out interface{}, isArray bool, err error) {
	t, err := time.Parse(TimeLayout, in)
	if err != nil {
		return
	}
	y, m, d := t.Date()
	out = y*1e4 + int(m)*1e2 + d
	return
}
