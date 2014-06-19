package mongosearch

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
	"testing"
	"time"
)

const ServerAddr = "192.168.20.15:49154/testdb"

var pubs = []bson.ObjectId{
	bson.ObjectIdHex("100000000000000000000000"),
	bson.ObjectIdHex("200000000000000000000000"),
	bson.ObjectIdHex("300000000000000000000000"),
}

func TestQuery(t *testing.T) {
	resetDB(t)

	s, err := New(ServerAddr, "Items", ServerAddr, "Results")
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.Search("a OR b")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("id: %s", id)
}

func resetDB(t *testing.T) {
	sess, err := mgo.Dial(ServerAddr)
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}
	defer sess.Close()

	d := sess.DB("")
	if _, err := d.C("Items").RemoveAll(nil); err != nil {
		t.Fatalf("Error dropping: %s", err)
	}

	k := func(str string) (kws []string) {
		words := strings.Fields(str)
		kws = make([]string, 0, len(words))
		for _, w := range words {
			if w < "0" || w > "9" {
				kws = append(kws, w)
			}
		}
		return
	}
	type Doc struct {
		Id    int
		PubId bson.ObjectId
		Date  time.Time
		All   []string
		Kws   []string `bson:"keywords"`
	}
	newDoc := func(id int, d string, text string) (doc Doc) {
		t, _ := time.Parse("2006-01-02", d)
		return Doc{
			Id:    id,
			PubId: pubs[id%len(pubs)],
			Date:  t,
			All:   strings.Fields(text),
			Kws:   k(text),
		}
	}
	docs := []interface{}{
		newDoc(1, "2014-06-01", "a 0 b 1 c 2 d"),
		newDoc(2, "2014-06-02", "a 0 1 b 0 c d e 2 f"),
		newDoc(3, "2014-06-02", "a 1 2 b 0 c e 2 g"),
		newDoc(4, "2014-06-03", "0 1 b 0 2"),
		newDoc(5, "2014-06-04", "a a b b c c"),
	}
	d.C("Items").Insert(docs...)
}
