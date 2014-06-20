package mongosearch

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

type Conversion func(string) (interface{}, bool, error)

var ConvertDate Conversion = func(in string) (out interface{}, isArray bool, err error) {
	out, err = time.Parse(TimeLayout, in)
	return
}

var ConvertBsonId Conversion = func(in string) (out interface{}, isArray bool, err error) {
	if !bson.IsObjectIdHex(in) {
		err = fmt.Errorf("Invalid BSON ObjectId: %s", in)
		return
	}
	out, err = bson.ObjectIdHex(in), nil
	return
}

var ConvertSpaces Conversion = func(in string) (out interface{}, isArray bool, err error) {
	isArray, out = true, strings.Fields(in)
	return
}