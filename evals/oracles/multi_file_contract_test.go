package fixture_test

import (
	"encoding/json"
	"fixture/api"
	"os"
	"reflect"
	"strings"
	"testing"
)

func evalPhoneField(t *testing.T, c *api.Customer) reflect.Value {
	t.Helper()
	f, ok := reflect.TypeOf(*c).FieldByName("Phone")
	if !ok || f.Type.Kind() != reflect.String {
		t.Fatal("Customer has no string field Phone as docs/FIELDS.md defines")
	}
	if tag := f.Tag.Get("json"); tag != "phone,omitempty" {
		t.Fatalf("Phone json tag = %q, want %q", tag, "phone,omitempty")
	}
	return reflect.ValueOf(c).Elem().FieldByName("Phone")
}

func evalCustomer(t *testing.T, phone string) api.Customer {
	t.Helper()
	c := api.Customer{ID: "cus_1", Email: "a@b.io", Name: "A"}
	evalPhoneField(t, &c).SetString(phone)
	return c
}

func TestEvalPhoneJSON(t *testing.T) {
	c := evalCustomer(t, "+4930123456")
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var back api.Customer
	if err := json.Unmarshal(data, &back); err != nil || !reflect.DeepEqual(back, c) {
		t.Fatalf("round trip = %+v, %v", back, err)
	}
	if !strings.Contains(string(data), `"phone":"+4930123456"`) {
		t.Fatalf("json = %s", data)
	}
	empty, _ := json.Marshal(evalCustomer(t, ""))
	if strings.Contains(string(empty), "phone") {
		t.Fatalf("empty phone is not omitted: %s", empty)
	}
}

func TestEvalPhoneValidation(t *testing.T) {
	valid := []string{"", "+4930123456", "+12025550123", "+12345678", "+123456789012345"}
	invalid := []string{"030123456", "4930123456", "+0301234567", "+1234567", "+1234567890123456", "+49 30 123456", "+4930-123456", "+", "++4930123456", "+49301234５6"}
	for _, p := range valid {
		if err := evalCustomer(t, p).Validate(); err != nil {
			t.Errorf("Validate(phone %q) = %v, want nil", p, err)
		}
	}
	for _, p := range invalid {
		if err := evalCustomer(t, p).Validate(); err == nil {
			t.Errorf("Validate(phone %q) = nil, want an error", p)
		}
	}
}

func TestEvalSampleGoldenAndChangelog(t *testing.T) {
	s := api.SampleCustomer()
	if evalPhoneField(t, &s).String() == "" {
		t.Error("SampleCustomer does not set Phone; CONTRIBUTING.md requires the sample to set every field")
	}
	if err := s.Validate(); err != nil {
		t.Errorf("SampleCustomer is invalid: %v", err)
	}
	got, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("api/testdata/customer.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(got)+"\n" != string(want) {
		t.Errorf("golden file does not match the sample:\n%s", got)
	}
	log, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	_, rest, _ := strings.Cut(string(log), "## Unreleased")
	unreleased, _, _ := strings.Cut(rest, "\n## ")
	if !strings.Contains(strings.ToLower(unreleased), "phone") {
		t.Error("CHANGELOG.md has no Unreleased entry for the phone field")
	}
}
