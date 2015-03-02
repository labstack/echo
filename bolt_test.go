package bolt

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"testing"
)

type (
	user struct {
		Id   string
		Name string
	}
)

var u = user{
	Id:   "1",
	Name: "Joe",
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
	if u2.Id != u.Id {
		t.Error("user id should be %s, found %s", u.Id, u2.Id)
	}
	if u2.Name != u.Name {
		t.Error("user name should be %s, found %s", u.Name, u2.Name)
	}
}
