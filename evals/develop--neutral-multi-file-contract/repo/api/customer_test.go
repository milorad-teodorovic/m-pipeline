package api

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGolden(t *testing.T) {
	got, err := json.MarshalIndent(SampleCustomer(), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/customer.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(got)+"\n" != string(want) {
		t.Fatalf("golden mismatch:\n%s", got)
	}
}

func TestSampleValid(t *testing.T) {
	if err := SampleCustomer().Validate(); err != nil {
		t.Fatal(err)
	}
}
