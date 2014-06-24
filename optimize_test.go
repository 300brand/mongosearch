package mongosearch

import (
	"github.com/300brand/searchquery"
	"testing"
)

var optimizeTests = []struct {
	In    string
	Query string
}{
	{
		`intdate:("2014-06-01 00:00:00" OR "2014-06-02 00:00:00" OR "2014-06-03 00:00:00") AND pubid:(52be3360b6bbac0ca102b8ac OR 528e455f84e7536d52001178) AND keywords:("monkey poo" OR "banana peel" OR zoo)`,
		`{
			"$and": [
				{
					"intdate": {
						"$in": [ 20140601, 20140602, 20140603 ]
					}
				},
				{
					"pubid": {
						"$in": [ "52be3360b6bbac0ca102b8ac", "528e455f84e7536d52001178" ]
					}
				},
				{
					"$or": [
						{ "keywords": { "$all": [ "monkey", "poo" ] } },
						{ "keywords": { "$all": [ "banana", "peel" ] } },
						{ "keywords": "zoo" }
					]
				}
			]
		}`,
	},
}

var canOptimizeTests = []struct {
	In       string
	Required bool
	Optional bool
}{
	{`space:(a OR b OR c)`, false, true},
	{`space:(a OR "b c" OR d)`, false, false},
	{`space:(a b c d)`, true, false},
	{`space:(a AND b AND c AND d)`, true, false},
	{`space:((a OR b) AND c AND d)`, false, false},
}

func TestOptimize(t *testing.T) {
	s, err := New("", "Items", "Results", "")
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}
	s.Rewrite("", "keywords")
	s.Convert("keywords", ConvertSpaces)
	s.Convert("intdate", ConvertDateInt)
	s.Convert("pubid", ConvertBsonId)

	for _, q := range optimizeTests {
		query, err := searchquery.ParseGreedy(q.In)
		if err != nil {
			t.Fatalf("Error parsing query: %s", err)
		}
		built, err := s.buildQuery(query)
		if err != nil {
			t.Fatalf("Error building query: %s - %s", query, err)
		}

		testResult(t, built, q.Query)
	}
}

func TestCanOptimize(t *testing.T) {
	s, _ := New("", "Items", "Results", "")
	s.Convert("space", ConvertSpaces)

	for _, test := range canOptimizeTests {
		query, err := searchquery.ParseGreedy(test.In)
		if err != nil {
			t.Fatalf("Error parsing query: %s", err)
		}

		subquery := query.Required[0].Query
		if s.canOptimize(subquery.Optional) != test.Optional {
			t.Errorf("Optional optimize, expected %v - %s", test.Optional, subquery.Optional)
		}
		if s.canOptimize(subquery.Required) != test.Required {
			t.Errorf("Required optimize, expected %v - %s", test.Required, subquery.Required)
		}
	}
}
