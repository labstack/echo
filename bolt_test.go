package bolt

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	user struct {
		ID   string
		Name string
	}
)

var u = user{
	ID:   "1",
	Name: "Joe",
}

func TestBoltMaxParam(t *testing.T) {
	b := New()
	b.SetMaxParam(8)
	if b.maxParam != 8 {
		t.Errorf("max param should be 8, found %d", b.maxParam)
	}
}

func TestBoltIndex(t *testing.T) {
	b := New()
	b.Index("example/public/index.html")
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	b.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestBoltStatic(t *testing.T) {
	b := New()
	b.Static("/js", "example/public/js")
	r, _ := http.NewRequest("GET", "/js/main.js", nil)
	w := httptest.NewRecorder()
	b.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func verifyUser(rd io.Reader, t *testing.T) {
	var l int64
	err := binary.Read(rd, binary.BigEndian, &l) // Body length
	if err != nil {
		t.Fatal(err)
	}
	bd := io.LimitReader(rd, l) // Body
	u2 := new(user)
	dec := json.NewDecoder(bd)
	err = dec.Decode(u2)
	if err != nil {
		t.Fatal(err)
	}
	if u2.ID != u.ID {
		t.Errorf("user id should be %s, found %s", u.ID, u2.ID)
	}
	if u2.Name != u.Name {
		t.Errorf("user name should be %s, found %s", u.Name, u2.Name)
	}
}
