package mongosearch

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestBuildQuery(t *testing.T) {
	for _, test := range tests {
		t.Logf("%s", test.Input)
		var dst bytes.Buffer
		json.Indent(&dst, []byte(test.KeywordQueries), "", "  ")
		t.Logf("%s", dst.Bytes())
	}
}
