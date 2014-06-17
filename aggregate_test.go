package mongosearch

import (
	"github.com/300brand/searchquery"
	"labix.org/v2/mgo/bson"
	"testing"
)

var aggregates = []struct {
	Query  string
	Expect bson.M
}{
	{
		"date>='2014-06-01 00:00:00' AND date<='2014-06-07 00:00:00' AND (a OR b)",
		bson.M{
			"$and": []bson.M{
				bson.M{"date": bson.M{"$gte": "2014-06-01 00:00:00"}},
				bson.M{"date": bson.M{"$lte": "2014-06-07 00:00:00"}},
				bson.M{
					"$or": []bson.M{
						bson.M{"keywords": bson.M{"$all": []string{"a"}}},
						bson.M{"keywords": bson.M{"$all": []string{"b"}}},
					},
				},
			},
		},
	},
}

func a() {
	_ = bson.M{
		"$and": []bson.M{
			bson.M{
				"date": bson.M{
					"$gte": nil,
				},
			},
			bson.M{
				"date": bson.M{
					"$lte": nil,
				},
			},
			bson.M{
				"$or": []bson.M{},
			},
		},
	}
}

func TestPrepareAggregate(t *testing.T) {
	s, err := New(ServerAddr, "Items", ServerAddr, "Results", "allwords", "date", "keywords")
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	for _, a := range aggregates {
		query, err := searchquery.ParseGreedy(a.Query)
		if err != nil {
			t.Fatalf("Error parsing query: %s", err)
		}
		q, err := s.buildQuery(query)
		if err != nil {
			t.Fatalf("Error building query: %s - %s", query, err)
		}
		t.Logf("%#v", q)
	}
}
