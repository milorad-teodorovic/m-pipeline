package customers

import (
	"encoding/json"
	"errors"
	"strings"
)

// CurrentVersion is the record version that Encode writes.
const CurrentVersion = 2

// ErrUnknownVersion reports a record version that this code cannot read.
var ErrUnknownVersion = errors.New("unknown customer record version")

type recordV1 struct {
	Version int    `json:"version"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}

type recordV2 struct {
	Version   int    `json:"version"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// Encode returns the stored JSON record for c.
func Encode(c Customer) ([]byte, error) {
	return json.Marshal(recordV2{Version: CurrentVersion, FirstName: c.FirstName, LastName: c.LastName, Email: c.Email})
}

// Decode reads a stored JSON record of any supported version.
func Decode(data []byte) (Customer, error) {
	var head struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return Customer{}, err
	}
	switch head.Version {
	case 1:
		var r recordV1
		if err := json.Unmarshal(data, &r); err != nil {
			return Customer{}, err
		}
		first, last, _ := strings.Cut(r.Name, " ")
		return Customer{FirstName: first, LastName: last, Email: r.Email}, nil
	case 2:
		var r recordV2
		if err := json.Unmarshal(data, &r); err != nil {
			return Customer{}, err
		}
		return Customer{FirstName: r.FirstName, LastName: r.LastName, Email: r.Email}, nil
	default:
		return Customer{}, ErrUnknownVersion
	}
}
