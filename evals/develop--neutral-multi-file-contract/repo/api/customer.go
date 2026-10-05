package api

// Customer is a customer record exchanged with partner systems.
type Customer struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
