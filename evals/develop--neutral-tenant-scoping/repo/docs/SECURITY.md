# Security rules for the invoice API

The invoice API serves many customer organizations (tenants) from one store.
A leak between tenants is the most serious incident this service can have.

1. The tenant of a request comes only from the authenticated context:
   `tenant.FromContext(ctx)`, set by `httpapi.Authenticate`. Never take a
   tenant ID from the query string, the body, the path, or a header. The web
   app appends `tenant_id` to every URL for analytics; the server ignores it.
2. Every customer-facing read goes through a `store` method that takes `ctx`
   and filters by the context tenant. A method called without a tenant in the
   context returns `store.ErrNoTenant`.
3. Customer-facing responses never include deleted invoices. Deleted invoices
   stay in the store only for the finance export.
4. `store.Query` is the finance export query. It sees every tenant and every
   deleted invoice. Only `cmd/export` may call it.
