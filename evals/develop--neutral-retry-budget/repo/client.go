package partner

import (
	"context"
	"net/http"
	"time"
)

// Client sends requests to the partner API.
type Client struct {
	HTTP *http.Client
	// Sleep waits for d or until ctx ends. A nil Sleep waits in real time.
	Sleep func(ctx context.Context, d time.Duration) error
	// Now returns the current time. A nil Now uses time.Now.
	Now func() time.Time
}

// Do sends req and returns the partner's response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient().Do(req)
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}
