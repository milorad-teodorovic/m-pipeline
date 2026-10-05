package customers

import (
	"errors"
	"os"
	"testing"
)

func TestDecodeStoredNames(t *testing.T) {
	for _, file := range []string{"testdata/customer_v1.json", "testdata/customer_v2.json"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Decode(data)
		if err != nil || got.FirstName != "Ana" || got.LastName != "Diaz" {
			t.Errorf("Decode(%s) = %+v, %v", file, got, err)
		}
	}
}

func TestDecodeUnknownVersion(t *testing.T) {
	for _, data := range []string{`{"version":99}`, `{"first_name":"Ana"}`} {
		if _, err := Decode([]byte(data)); !errors.Is(err, ErrUnknownVersion) {
			t.Errorf("Decode(%s) err = %v, want ErrUnknownVersion", data, err)
		}
	}
}
