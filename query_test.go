package mongosearch

import (
	"bytes"
	"encoding/json"
	"github.com/300brand/searchquery"
	"testing"
)

func TestBuildQuery(t *testing.T) {
	for i, test := range tests {
		ms, _ := New("", "Items", "Results")
		ms.SetAll("text.words.all")
		ms.SetKeyword("text.words.keywords", ConvertSpaces, "keywords")
		ms.SetPubdate("pubdate.date", ConvertDateInt, "date", "pubdate", "published")
		ms.SetPubid("publicationid", ConvertBsonId, "pubid")

		query, err := searchquery.ParseGreedy(test.Input)
		if err != nil {
			t.Fatalf("searchquery.ParseGreedy: %s", err)
		}

		// Test fields
		fields := ms.mapFields(query)

		// Test reduce
		reduced := ms.reduce(fields[ms.fields.keyword])
		if rstr := reduced.String(); rstr != test.Reduced {
			t.Errorf("[%d] Reduced did not match", i)
			t.Errorf("[%d] Expect: %s", i, test.Reduced)
			t.Errorf("[%d] Got:    %s", i, rstr)
		}

		mgoQuery, err := ms.buildQuery(query)
		if err != nil {
			t.Fatalf("mongosearch.buildQuery: %s", err)
		}

		b, err := json.MarshalIndent(mgoQuery, "", "  ")
		if err != nil {
			t.Fatalf("json.MarshalIndent: %s", err)
		}

		if test.MapReduce != ms.reqMapReduce {
			t.Errorf("[%d] Map Reduce flag did not match", i)
			t.Errorf("[%d] Query: %s", i, test.Input)
			t.Errorf("[%d] Expected: %v", i, test.MapReduce)
		}

		// t.Logf("mgoQuery: %s", b)

		// t.Logf("%s", test.Input)
		var dst bytes.Buffer
		// json.Indent(&dst, []byte(test.KeywordQueries), "", "  ")
		// t.Logf("%s", dst.Bytes())
		_, _ = b, dst
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		Input, Reduced string
	}{
		{
			`keywords:a`,
			`keywords:a`,
		},
		{
			`keywords:(a)`,
			`+keywords:a`,
		},
		{
			`keywords:(a OR b)`,
			`keywords:a keywords:b`,
		},
		{
			`keywords:((a OR b))`,
			`keywords:a keywords:b`,
		},
		{
			`keywords:(a OR (b AND c))`,
			`keywords:a (+keywords:b +keywords:c)`,
		},
		{
			`keywords:((a OR b) AND c)`,
			`+(keywords:a keywords:b) +keywords:c`,
		},
		{
			`keywords:((a OR b OR c) NOT (d OR e))`,
			`keywords:a keywords:b keywords:c`,
		},
		{
			`keywords:((a) OR (b))`,
			`keywords:a keywords:b`,
		},
	}

	ms, _ := New("", "Items", "Results")
	for i, test := range tests {
		q, _ := searchquery.ParseGreedy(test.Input)
		fields := ms.mapFields(q)
		reduced := ms.reduce(fields["keywords"])
		if reduced.String() != test.Reduced {
			t.Errorf("[%d] Input:  %s", i, test.Input)
			t.Errorf("[%d] Expect: %s", i, test.Reduced)
			t.Errorf("[%d] Got:    %s", i, reduced)
			b, _ := json.MarshalIndent(reduced, "", "  ")
			t.Errorf("[%d] JSON:\n%s", i, b)
		}
	}
}
