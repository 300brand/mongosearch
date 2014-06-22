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
				{"date": {"$gte": "2014-06-01T00:00:00Z"}},
				{"date": {"$lte": "2014-06-07T00:00:00Z"}},
				{
					"$or": [
						{"keywords": {"$all": ["a"]}},
						{"keywords": {"$all": ["b"]}}
					]
				}
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
				{
					"$or": [
						{"keywords": {"$all": ["CDW"]}},
						{"keywords": {"$all": ["CDW-G"]}},
						{"keywords": {"$all": ["CDWG"]}}
					]
				}
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
				{
					"$or": [
						{"intdate": 20140601},
						{"intdate": 20140602},
						{"intdate": 20140603}
					]
				}
			]
		}`,
		`{ "and": [ { "or": [] } ] }`,
	},
}

func TestBuild(t *testing.T) {
	s, err := New("", "Items", "Results", "")
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}
	s.Rewrite("", "keywords")
	s.Convert("keywords", ConvertSpaces)
	s.Convert("date", ConvertDate)
	s.Convert("intdate", ConvertDateInt)
	s.Convert("pubid", ConvertBsonId)

	for _, q := range queries {
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

		testResult := func(in interface{}, expected string) {
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
			got := buf.Bytes()
			if !bytes.Equal(expect, got) {
				t.Logf("Does not match expected")
				t.Logf("Expected: %s", expect)
				t.Logf("Got:      %s", got)
				t.Fail()
			}
		}
		testResult(built, q.Query)
		testResult(scope, q.Scope)
	}
}
