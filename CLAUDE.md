# Project Rules

## Releasing

Releases are fully automated. Any push to `main` that changes Go source,
`go.mod`, `gemini-extension.json`, or `GEMINI.md` triggers the GitHub Actions
release workflow, which auto-increments the patch version, builds cross-platform
binaries, and creates a GitHub Release.

No manual tagging or version bumping is needed.
