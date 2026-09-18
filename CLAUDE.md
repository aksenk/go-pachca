# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`go-pachca` — Go client library for the [Пачка](https://pachca.com) messenger API (`https://api.pachca.com/api/shared/v1`). Module: `github.com/aksenk/go-pachca`. No `main` package — this is a library only. README.md tracks per-endpoint implementation status; update it when adding/finishing an endpoint.

## Commands

```bash
go build ./...                  # build
go test ./...                   # test
go test -v -race ./...          # test as run by pre-commit
go test -run TestChats_New_Success ./...   # single test
go vet ./...
gofmt -l .                      # list unformatted files (pre-commit uses this, not -w)
goimports -l .
staticcheck ./...
golangci-lint run
```

Pre-commit hooks (lefthook, `.lefthook.yml`) run gofmt, govet, gotest (`-race`), staticcheck, goimports, and golangci-lint in parallel. Run the relevant command yourself before committing if unsure.

## Architecture

The `Client` (`client.go`) is a struct of sub-resource clients, each wrapping the same shared `*resty.Client`:

```go
Client{ Messages, Threads, Users, Chats, Tags, Reactions, Files, Views, Profile }
```

`NewClient(*ClientOptions)` builds one `resty.Client` with base URL, bearer auth, and retry behavior (retries on network timeouts, `io.EOF`/connection reset, HTTP 429, and 5xx), then hands that same client to every sub-resource. `ClientOptions.RetryObserver` is invoked on every retry with `RetryMeta` for caller-side instrumentation/logging.

Each API area lives in its own file (`chats.go`, `chat_members.go`, `messages.go`, `users.go`, `tags.go`, `reactions.go`, `threads.go`, `files.go`, `views.go`, `profile.go`, `webhooks.go`) and follows the same per-method shape:

```go
func (x *X) Method(ctx context.Context, ...) (*XResponse, *resty.Response, error) {
    // 1. validate input params -> wrap ErrInvalidInput
    // 2. build URL from the *URL const in client.go
    // 3. client.R().SetContext(ctx)...Get/Post/Put(url)
    // 4. check resp.StatusCode() against the expected success code -> wrap ErrResponseCode
    // 5. json.Unmarshal into a *Raw wrapper type -> wrap ErrResponseDecode
    // 6. return &raw.Data, resp, nil
}
```

Every exported method returns `(result, *resty.Response, error)` (or `(*resty.Response, error)` for actions with no response body) — the raw `*resty.Response` is always returned even on error so callers can inspect status/headers. Sentinel errors (`ErrResponseCode`, `ErrResponseDecode`, `ErrInvalidInput`, etc., in `client.go`) are wrapped with `fmt.Errorf("%w: ...")`, so check errors with `errors.Is`.

Two response-shape conventions:
- Single-object endpoints unmarshal into a `<Type>ResponseRaw{ Data <Type>Response }` wrapper.
- `<Type>Response` structs embed the plain request `<Type>` struct plus server-assigned/readonly fields (e.g. `ChatResponse` embeds `Chat`; `UserResponse` embeds `User`).

**Pagination** — the API has two incompatible schemes in use across resources:
- Page/per-page (`PaginationOptions{Per, Page}`): used by `Chats.List`/`Find` and the deprecated `Users.List`/`getUsersPaginated`. Loop increments `Page` until a short page is returned.
- Cursor-based (`PaginationOptionsUsers{Limit, Next}`, cursor returned in `meta.paginate.next_page`): the current API for users — `Users.ListV2`/`Find`/`getUsersPaginatedV2`. Prefer this over the `V1`/deprecated methods when adding new user-listing code; the docstrings on the V1 methods explain why they're deprecated (per-page pagination doesn't work correctly on that endpoint).

Each `List`/`Find`-style method loops over its own private `get<X>Paginated` helper, accumulating results until a short/empty page (or empty cursor) signals the end.

**Caching** — `Users` and `Chats` each hold `cache map[int]*XResponse` + `*sync.RWMutex`, populated lazily via `GetWithCache`. The cache is unbounded and never invalidated/expired by `Update`/`Delete`-style calls — keep that in mind when adding mutating methods for cached resources.

## Testing conventions

Tests mock at the transport level: build a real `resty.New()` and swap in a `mockTransport` (`RoundTrip` returning a fixed `*http.Response`) — see `chats_test.go`. There's no interface/mock for the `Chats`/`Users`/etc. structs themselves; instantiate them directly with `&Chats{client: client}` (cache/mutex fields can be left nil for tests that don't exercise `GetWithCache`).

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
