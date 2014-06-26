package mongosearch

import (
	"time"
)

var today = time.Now().Format(TimeLayout)

var tests = []struct {
	Input          string
	KeywordQueries string
	FullQuery      string
	Scope          string
}{
	{
		`a OR b`,
		`[
			{ "kw": "a" },
			{ "kw": "b" }
		]`,
		`{
			"$or": [
				{
					"pubdate": { "$in": [ ` + today + ` ] },
					"kw": "a"
				},
				{
					"pubdate": { "$in": [ ` + today + ` ] },
					"kw": "b"
				}
			]
		}`,
		`{}`,
	},
	{
		`a OR "b c"`,
		`[
			{ "kw": "a"},
			{ "kw": { "$all": [ "b", "c" ] } }
		]`,
		`{
			"$or": [
				{
					"pubdate": { "$in": [ ` + today + ` ] },
					"kw": "a"
				},
				{
					"pubdate": { "$in": [ ` + today + ` ] },
					"kw": { "all": [ "b", "c" ] }
				}
			]
		}`,
		`{}`,
	},
}
