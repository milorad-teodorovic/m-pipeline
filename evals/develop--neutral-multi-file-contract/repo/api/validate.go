package api

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// Validate reports the first field of c that breaks docs/FIELDS.md.
func (c Customer) Validate() error {
	if !validID(c.ID) {
		return errors.New("invalid id")
	}
	if strings.Count(c.Email, "@") != 1 || strings.HasPrefix(c.Email, "@") || strings.HasSuffix(c.Email, "@") {
		return errors.New("invalid email")
	}
	if n := utf8.RuneCountInString(c.Name); n < 1 || n > 80 {
		return errors.New("invalid name")
	}
	return nil
}

func validID(id string) bool {
	rest, ok := strings.CutPrefix(id, "cus_")
	if !ok || rest == "" {
		return false
	}
	for _, r := range rest {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
