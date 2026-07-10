# Development Log

This log records the main development decisions, implementation steps, and verification results for the forum project. Each contributor should add a new entry when they introduce a feature, change the architecture, adjust data models, or fix an important bug.

## Entry Format

Use this structure for every new entry:

```md
## Day N - Short Title

**Date:** YYYY-MM-DD
**Author:** Name
**Branch:** branch-name

### Goal
Briefly explain what the work was meant to achieve.

### Implementation
- Describe the files or packages changed.
- Explain the important design decisions.
- Mention any constraints or assumptions.

### Verification
- List commands run, manual checks done, or tests added.
- Note anything that could not be tested.

### Next Steps
- List follow-up work that should happen after this entry.
```

Keep entries short but useful. The goal is not to write a diary; the goal is to help the next developer understand what changed, why it changed, and how to continue without guessing.

## Day 1 - Go Data Definitions

**Date:** 2026-07-10

**Author:** [Bramwel Mutugi](https://learn.zone01kisumu.ke/git/mumutugi)

**Branch:** `feature/domain-structs`

### Goal

Set up the first Go layer of the project using pure data definitions only. This branch intentionally avoids application behavior, database queries, HTTP handlers, and business logic so the rest of the team can build on stable shared types.

### Implementation

- Added a minimal `go.mod` with module name `forum` so all packages can import each other consistently.
- Defined configuration structs in `internal/config/config.go`, including application, server, database, session, and security settings.
- Defined core domain models in `internal/domain`, including users, sessions, posts, categories, comments, votes, stats, filters, and draft/input models.
- Defined handler-facing request and view data structs in `internal/handler`, such as auth forms, post views, comment forms, vote requests, pagination data, flash messages, and request context.
- Defined repository-facing data containers in `internal/repository`, keeping them separate from HTTP concerns.
- Defined SQLite row structs in `internal/repository/sqlite` for users, sessions, posts, categories, comments, votes, migrations, and aggregate stats.
- Added small application/dependency structs in `cmd/forum/main.go` to reserve the future composition shape without starting the runtime application yet.

After reviewing the project instructions, the data models were refined to support the required forum behavior:

- Added `PostFilterKind` in `internal/domain/post.go` with `all`, `category`, `created`, and `liked` filter modes.
- Added `ViewerID` to `PostFilter` so created-post and liked-post filters can be scoped to the logged-in user.
- Added `VoteNone` to represent users who have not liked or disliked a post or comment.
- Added `UserVote` to post/comment view models and repository records so public pages can show total likes/dislikes while logged-in users can also see their own vote state.
- Added `UserVote` fields to SQLite stats rows for posts and comments so repository queries can return aggregate counts and the current user's vote together.

The main design decision was to separate data definitions by project layer:

- `internal/domain` holds business entities and shared domain types.
- `internal/config` holds runtime configuration shapes.
- `internal/handler` holds HTTP request and view payload shapes.
- `internal/repository` holds persistence-facing records and query shapes.
- `internal/repository/sqlite` holds database row shapes specific to SQLite.

This keeps the project ready for implementation while avoiding early coupling between handlers, repositories, and domain logic.

### Verification

Formatted the Go code with `gofmt`.

Compiled all packages with:

```sh
env GOCACHE=/tmp/go-build-cache go test -v ./...
```

The command passed across all current packages. There are no test files yet because this entry only introduces data structures.

### Next Steps

- Add the database schema that matches the SQLite row definitions.
- Implement repository interfaces and SQLite queries.
- Implement HTTP handlers using the handler request/view structs.
- Add validation rules for user registration, login, posts, comments, and votes.
- Add tests once behavior is introduced.
