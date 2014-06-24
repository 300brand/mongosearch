package mongosearch

import (
	"bytes"
	"encoding/json"
	"github.com/300brand/searchquery"
	"strings"
	"testing"
)

var queries = []struct {
	In    string
	Query string
	Scope string
}{
	{
		"date>='2014-06-01 00:00:00' AND date<='2014-06-07 00:00:00' AND (a OR b)",
		`{
			"$and": [
				{ "date": { "$gte": "2014-06-01T00:00:00Z" } },
				{ "date": { "$lte": "2014-06-07T00:00:00Z" } },
				{ "keywords": { "$in": [ "a", "b" ] } }
			]
		}`,
		`{
			"and": [
				{
					"or": [ "a", "b" ]
				}
			]
		}`,
	},
	{
		"date>='2014-06-01 00:00:00' AND date<='2014-06-07 00:00:00' AND ('a 0 b')",
		`{
			"$and": [
				{"date": {"$gte": "2014-06-01T00:00:00Z"}},
				{"date": {"$lte": "2014-06-07T00:00:00Z"}},
				{
					"$and": [
						{"keywords": {"$all": ["a", "0", "b"]}}
					]
				}
			]
		}`,
		`{
			"and": [
				{
					"and": [ "a 0 b" ]
				}
			]
		}`,
	},
	{
		"('CDW' OR 'CDW-G' OR 'CDWG') NOT ('collision damage waiver')",
		`{
			"$and": [
				{ "keywords": { "$in": [ "CDW", "CDW-G", "CDWG" ] } }
			]
		}`,
		`{
			"and": [
				{
					"or": [ "CDW", "CDW-G", "CDWG" ]
				}
			],
			"nor": [
				{
					"and": [ "collision damage waiver" ]
				}
			]
		}`,
	},
	{
		"intdate:('2014-06-01 00:00:00' OR '2014-06-02 08:00:00' OR '2014-06-03 05:00:00')",
		`{
			"$and": [
				{ "intdate": { "$in": [ 20140601, 20140602, 20140603 ] } }
			]
		}`,
		`{ "and": [ { "or": [] } ] }`,
	},
	{
		`intdate:('2014-06-01 00:00:00' OR '2014-06-02 00:00:00') AND ("monkey" AND "banana") AND pubid:(53678fb4800b8e4c9d0002c9 OR 53678ea54113de7739000214 OR 53678ea54113de7739000211)`,
		`{
			"$and": [
				{ "intdate": { "$in": [ 20140601, 20140602 ] } },
				{ "keywords": { "$all": [ "monkey", "banana" ] } },
				{
					"pubid": {
						"$in": [
							"53678fb4800b8e4c9d0002c9",
							"53678ea54113de7739000214",
							"53678ea54113de7739000211"
						]
					}
				}
			]
		}`,
		`{"and":[{"or":[]},{"and":["monkey","banana"]},{"or":[]}]}`,
	},
	{
		`"data center" AND "Google"`,
		`{
			"$and": [
				{ "keywords": { "$all": [ "data", "center" ] } },
				{ "keywords": "Google" }
			]
		}`,
		`{
			"and": [
				"data center",
				"Google"
			]
		}`,
	},
}

func TestBuild(t *testing.T) {
	for _, q := range queries {
		s, err := New("", "Items", "Results", "")
		if err != nil {
			t.Fatalf("Error connecting: %s", err)
		}
		s.Rewrite("", "keywords")
		s.Convert("keywords", ConvertSpaces)
		s.Convert("date", ConvertDate)
		s.Convert("intdate", ConvertDateInt)
		s.Convert("pubid", ConvertBsonId)

		query, err := searchquery.ParseGreedy(q.In)
		if err != nil {
			t.Fatalf("Error parsing query: %s", err)
		}
		built, err := s.buildQuery(query)
		if err != nil {
			t.Fatalf("Error building query: %s - %s", query, err)
		}
		scope, err := s.buildScope(query)
		if err != nil {
			t.Fatalf("Error building scope: %s - %s", query, err)
		}

		testResult(t, built, q.Query)
		testResult(t, scope, q.Scope)
	}
}

func testResult(t *testing.T, in interface{}, expected string) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	dec := json.NewDecoder(strings.NewReader(expected))
	dec.UseNumber()

	var v interface{}
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("Error decoding expected result: %s\n%s", err, expected)
	}
	if err := enc.Encode(&v); err != nil {
		t.Fatalf("Error re-encoding expected result: %s", err)
	}
	expect := make([]byte, buf.Len())
	copy(expect, buf.Bytes())
	buf.Reset()
	if err := enc.Encode(&in); err != nil {
		t.Fatalf("Error encoding generated value: %s", err)
	}
	got := make([]byte, buf.Len())
	copy(got, buf.Bytes())
	if !bytes.Equal(expect, got) {
		t.Logf("Does not match expected")
		buf.Reset()
		json.Indent(&buf, expect, "", "\t")
		t.Logf("Expected:\n%s", buf.Bytes())
		buf.Reset()
		json.Indent(&buf, got, "", "\t")
		t.Logf("Got:\n%s", buf.Bytes())
		t.Fail()
	}
}
