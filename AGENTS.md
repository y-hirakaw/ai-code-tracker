# Repository Guidelines

AIエージェントの返答は日本語で行うこと

## Project Structure & Module Organization
- `cmd/aict`: CLI entrypoint (main package). Build targets live here.
- `internal/*`: Core packages: `tracker`, `storage`, `branch`, `period`, `git`, `validation`, `templates`, `security`, `errors`.
- `benchmarks/`: Performance tests; `docs/`: design notes and improvements.
- Tests live beside code as `*_test.go` (including `integration_test.go`).

## Build, Test, and Development Commands
- Build CLI: `go build -o bin/aict ./cmd/aict`
- Run locally: `go run ./cmd/aict --help`
- All tests: `go test ./...`
- With coverage: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Static checks: `go vet ./...`; formatting: `go fmt ./...`

## Coding Style & Naming Conventions
- Go defaults: tabs for indentation; 100–120 col soft limit.
- Package names: short, lowercase (e.g., `tracker`, `period`).
- Files: feature-oriented (e.g., `analyzer.go`, `filter.go`); tests mirror names with `_test.go`.
- Public APIs use clear nouns/verbs; avoid stutter (`tracker.Analyzer`, not `tracker.TrackerAnalyzer`).
- Keep CLI flags explicit and consistent across subcommands.

## Testing Guidelines
- Unit tests co-located with packages; integration in root when spanning modules.
- Name tests by subject and behavior (e.g., `TestAnalyzer_JSONL_RoundTrip`).
- Aim for meaningful coverage on `internal/tracker`, `internal/storage`, and `internal/branch` logic.
- Use table-driven tests; prefer deterministic fixtures over time-based assertions.

## Commit & Pull Request Guidelines
- Follow Conventional Commits: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, etc.
- PRs include: concise summary, linked issues, before/after notes, and CLI examples if UX changes.
- Require: `go test ./...` green; `go vet` clean; updated docs when flags/outputs change.

## Agent Tips (AI Contributors)
- Keep changes minimal and focused; prefer small PRs.
- Update `README.md` and `docs/` when modifying commands, flags, or outputs.
- Do not add new tools unless required; rely on `go build/test/vet/fmt`.
- Preserve existing directory boundaries; avoid cross-package leaks.
