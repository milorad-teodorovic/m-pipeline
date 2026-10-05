package tenant

import "context"

type key struct{}

// WithTenant returns a context that carries the tenant ID.
func WithTenant(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key{}, id)
}

// FromContext returns the tenant ID in ctx and whether one is set.
func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(key{}).(string)
	return id, ok && id != ""
}
