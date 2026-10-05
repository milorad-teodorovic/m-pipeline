package main

import (
	"encoding/csv"
	"os"
	"strconv"

	"fixture/store"
)

func main() {
	s := store.New()
	w := csv.NewWriter(os.Stdout)
	for _, inv := range s.Query(store.Filter{}) {
		_ = w.Write([]string{inv.TenantID, inv.ID, inv.Status, strconv.FormatInt(inv.Cents, 10), strconv.FormatBool(inv.Deleted)})
	}
	w.Flush()
}
