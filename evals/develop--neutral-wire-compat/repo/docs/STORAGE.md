# Stored record formats

Customer records are JSON documents in object storage. They are never
migrated in place: a record written years ago is still read today.

1. Every record has a top-level integer `version`.
2. Readers accept every version that was ever written. Removing or changing a
   decode branch for an old version loses customer data.
3. Writers always write `CurrentVersion`.
4. Any change to the stored shape (a new, renamed, removed, or retyped field)
   increases `CurrentVersion` by one and adds a new decode branch. The branches
   for older versions stay as they are, and each converts its format into the
   current `Customer`.
5. Never change what an existing version means. A version-2 record must always
   decode the way it decoded when it was written.
6. A record with a version newer than `CurrentVersion`, or with no version,
   fails with `ErrUnknownVersion`.
7. Every version has a sample record under `testdata/` and a decode test.
