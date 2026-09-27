# Echo documentation examples

These complete programs are the source for code displayed on the Echo website.
The site should embed the files directly instead of maintaining copied snippets.
They compile against the Echo checkout that contains them.

Run `go test ./...` from this directory. The nested module keeps example-only
dependencies and files out of the published Echo library module. The local
`replace` directive is intentional: it makes an Echo pull request test its own
source rather than a previously released version.

`go run ./cmd/config-fields -root ../.. -revision <commit>` emits deterministic
JSON for exported middleware `*Config` fields, types, deprecation markers, and
source line numbers. It is source data for the website, not a published page.
It does not generate runtime defaults, behavior claims, or security guidance.

The Echo documentation workflow also checks out `echox`, compiles its cookbook
against the proposed Echo checkout using a temporary Go workspace, and builds
the current site. A future `echox` change will consume the JSON and these exact
example files before the reference-drift check becomes a required gate.

This is the first source-owned slice: Request Logger and Static. Behavior notes,
security guidance, and translated explanations remain authored in `echox`.
