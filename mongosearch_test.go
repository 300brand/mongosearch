package mongosearch

import (
	"labix.org/v2/mgo"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const ServerAddr = "192.168.20.15:49154"

// const ServerAddr = "squeaker"

func xTestQuery(t *testing.T) {
	db := "testdb"
	resetDB(t, db)

	addr := filepath.Join(ServerAddr, db)
	s, err := New(addr, "Items", addr, "Results", "all", "date", "keywords")
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.Search("a OR b")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("id: %s", id)
}

func resetDB(t *testing.T, db string) {
	sess, err := mgo.Dial(ServerAddr + "/" + db)
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
		Id   int
		Date time.Time
		All  []string
		Kws  []string `bson:"keywords"`
	}
	newDoc := func(id int, d string, text string) (doc Doc) {
		t, _ := time.Parse("2006-01-02", d)
		return Doc{
			Id:   id,
			Date: t,
			All:  strings.Fields(text),
			Kws:  k(text),
		}
	}
	docs := []interface{}{
		newDoc(1, "2014-06-01", "a 0 b 1 c 2 d"),
	}
	d.C("Items").Insert(docs...)
}
