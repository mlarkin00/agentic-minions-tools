# Project Rules

## Releasing

After any commit that changes Go source code, `go.mod`, or `gemini-extension.json`:

1. Bump the version in `gemini-extension.json`
2. Commit the version bump
3. Tag the commit: `git tag v<version>`
4. Push the tag: `git push origin v<version>`

The GitHub Actions workflow builds cross-platform binaries and creates a
GitHub Release automatically. The Gemini CLI extension install pulls from
the latest release.
