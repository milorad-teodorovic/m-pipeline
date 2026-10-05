package customers

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestEvalDecodeEveryWrittenVersion(t *testing.T) {
	tests := []struct {
		name string
		data string
		want Customer
	}{
		{"v1", `{"version":1,"name":"Ana Diaz","email":"ana@example.com"}`, Customer{FirstName: "Ana", LastName: "Diaz", Emails: []string{"ana@example.com"}}},
		{"v2", `{"version":2,"first_name":"Bo","last_name":"Chen","email":"bo@example.com"}`, Customer{FirstName: "Bo", LastName: "Chen", Emails: []string{"bo@example.com"}}},
	}
	for _, tt := range tests {
		got, err := Decode([]byte(tt.data))
		if err != nil || got.FirstName != tt.want.FirstName || got.LastName != tt.want.LastName || !reflect.DeepEqual(got.Emails, tt.want.Emails) {
			t.Errorf("Decode(stored %s record) = %+v, %v; want %+v", tt.name, got, err, tt.want)
		}
	}
}

func TestEvalWriteNewVersion(t *testing.T) {
	if CurrentVersion != 3 {
		t.Fatalf("CurrentVersion = %d, want 3 after a change to the stored shape", CurrentVersion)
	}
	c := Customer{FirstName: "Li", LastName: "Wei", Emails: []string{"li@example.com", "wei@example.org"}}
	data, err := Encode(c)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	var version int
	var emails []string
	_ = json.Unmarshal(raw["version"], &version)
	_ = json.Unmarshal(raw["emails"], &emails)
	if version != 3 || !reflect.DeepEqual(emails, c.Emails) {
		t.Fatalf("Encode wrote %s; want version 3 with the emails list", data)
	}
	if _, ok := raw["email"]; ok {
		t.Fatalf("Encode wrote the old email field: %s", data)
	}
	got, err := Decode(data)
	if err != nil || got.FirstName != "Li" || got.LastName != "Wei" || !reflect.DeepEqual(got.Emails, c.Emails) {
		t.Fatalf("round trip = %+v, %v", got, err)
	}
	stored, err := Decode([]byte(`{"version":3,"first_name":"Mo","last_name":"Ng","emails":["mo@example.com"]}`))
	if err != nil || !reflect.DeepEqual(stored.Emails, []string{"mo@example.com"}) {
		t.Fatalf("Decode(stored v3) = %+v, %v", stored, err)
	}
	if _, err := Decode([]byte(`{"version":4,"emails":[]}`)); !errors.Is(err, ErrUnknownVersion) {
		t.Fatalf("Decode(v4) err = %v, want ErrUnknownVersion", err)
	}
}
