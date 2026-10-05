# Contributing

## Changing a public API type

The types in `api/` are a public contract. Partner systems parse our JSON.
When you add a field to a public type, do all of these in the same change:

1. Name the Go field and its JSON key as `docs/FIELDS.md` defines them.
   Optional fields use `omitempty`.
2. Enforce every constraint from `docs/FIELDS.md` in the type's `Validate`
   method.
3. Set the field in the type's sample constructor in `api/sample.go`. The
   sample must set every field and must pass `Validate`.
4. Regenerate the golden file under `api/testdata/` so that it matches the
   sample. The golden file is the indented JSON of the sample, as produced by
   `json.MarshalIndent(v, "", "  ")`, followed by one newline.
5. Add a line under `## Unreleased` in `CHANGELOG.md`.
