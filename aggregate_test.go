package mongosearch

import (
	"bytes"
	"encoding/json"
	"github.com/300brand/searchquery"
	"strings"
	"testing"
)

var aggregates = []struct {
	Query  string
	Expect string
}{
	{
		"date>='2014-06-01 00:00:00' AND date<='2014-06-07 00:00:00' AND keywords:(a OR b)",
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
	},
}

func TestPrepareAggregate(t *testing.T) {
	s, err := New(ServerAddr, "Items", ServerAddr, "Results")
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}
	s.Rewrite("", "keywords")
	s.Convert("keywords", ConvertSpaces)
	s.Convert("date", ConvertDate)
	s.Convert("pubid", ConvertBsonId)

	for _, a := range aggregates {
		query, err := searchquery.ParseGreedy(a.Query)
		if err != nil {
			t.Fatalf("Error parsing query: %s", err)
		}
		q, err := s.buildQuery(query)
		if err != nil {
			t.Fatalf("Error building query: %s - %s", query, err)
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		dec := json.NewDecoder(strings.NewReader(a.Expect))
		dec.UseNumber()

		var v interface{}
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("Error decoding expected result: %s\n%s", err, a.Expect)
		}
		if err := enc.Encode(&v); err != nil {
			t.Fatalf("Error re-encoding expected result: %s", err)
		}
		expect := make([]byte, buf.Len())
		copy(expect, buf.Bytes())
		buf.Reset()
		if err := enc.Encode(&q); err != nil {
			t.Fatalf("Error encoding generated query: %s", err)
		}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Logf("Query does not match expected")
			t.Logf("Expected: %s", expect)
			t.Logf("Got:      %s", got)
			t.Fail()
		}
	}
}
