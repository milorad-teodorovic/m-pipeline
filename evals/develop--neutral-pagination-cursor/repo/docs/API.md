# Catalog HTTP API

Owner: API guild. Every list endpoint follows the rules in this file.

## Response shape

A list response is a JSON object:

    {"items": [ ... ], "next_cursor": "..."}

`next_cursor` is an empty string on the last page. Clients stop when they
receive an empty `next_cursor`, so it must be empty exactly when no item
remains after the page.

## Ordering

Items are ordered by `created_at` ascending, then by `id` ascending. Many
items share a `created_at` value (bulk imports), so the `id` tie-break is
required for a stable order.

## Page size

The `limit` query parameter sets the page size. A missing, empty, or zero
`limit` means 20. A `limit` above 100 is clamped to 100; it is not an error.
A negative or non-numeric `limit` is a 400 Bad Request.

## Cursors

Cursor-based lists take the `cursor` query parameter. A cursor is opaque to
clients: they never build or parse one.

A cursor must stay valid while the data changes. Items are created and
deleted between two page requests all the time. When a client walks the list
page by page:

- it never receives the same item twice,
- it never misses an item that existed for the whole walk,
- a deleted item simply stops appearing.

A cursor that the server cannot decode is a 400 Bad Request.
