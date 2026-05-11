# Go Code Convention

## Project structure

They generally follow the same style of project structure as follows:

```
.github/              # GitHub-specific configuration (CI/CD workflows, templates, automation)

cmd/                  # Application entry points (main packages that build binaries)
    <api>/
        main.go       # Entry point of the API service

configs/              # Configuration files (env, yaml, json, system configs)

db/                   # Database-related components
    migrations/       # Database schema migrations (versioned changes)

docs/                 # Project documentation (API docs, architecture, design docs)

internal/             # Core business logic (Go internal package - cannot be imported externally)
    handlers/         # HTTP request handlers (controller layer)
    middlewares/      # Middleware components (auth, logging, rate limiting, etc.)
    models/           # Data models / domain structs representing database entities
    repositories/     # Data access layer (database queries, ORM logic)
    routes/           # Route definitions and mapping to handlers
    services/         # Business logic layer (core application rules)
    worker/           # Background workers

pkg/                  # Shared libraries that can be reused across projects
```

## Import packages

All package import paths must be valid Go Modules path except packages from standard library.

Basic rules:

- Use different groups to separate import paths by an blank line for two or more types of packages.

Here is a concrete example:

```go

import (
    "fmt"

    "html/template"
    "net/http"

    "os"

    "github.com/urfave/cli"

    "gopkg.in/macaron.v1"

    "github.com/gogs/git-module"

    "github.com/gogs/gogs/internal/route"

    "github.com/gogs/gogs/internal/route/repo"

    "github.com/gogs/gogs/internal/route/user"
)
```

## Comment

- If the object is countable and does not know the number of it, use singular form and present tense; otherwise, use plural form.

- Maximum length of a comment line should be 80 characters.

## Naming rules

### Files
- All file names must be written in English.
- Use `snake_case` for all file names (lowercase letters with underscores only).
- The file contains the entry point of the application or package should be named as `main.go` or same as the package.

### Struct and interface

- Struct and interface names must follow standard English words or widely accepted abbreviations.

- Use PascalCase for struct and interface names.

- Struct and interface names should be in singular form.

```go

  type Webhook struct {

    ID           int64

    RepoID       int64

    OrgID        int64

    URL          string

    ContentType  HookContentType

    Secret       string

    Events       string

    *HookEvent

    IsSSL        bool

    IsActive     bool

    HookTaskType HookTaskType

    Meta         string

    LastStatus   HookStatus

    CreatedAt    time.Time

    UpdatedAt    time.Time

  }

```

- Type is often described in singular form:

  ```Go

  type Request struct { ...

  ```

- Interface should be described as follows:

  ```Go

  type FileInfo interface { ...

  ```

### Functions and methods

- A function and methods name should follow general English expression or shorthand.

- Use camelCase

- A function and methods name should singular form

- If the main purpose of the function or method is returning a `bool` type value, the name of function or method should starts with `Has`, `Is`, `Can` or `Allow`, etc.

  ```go

  func HasPrefix(name string, prefixes []string) bool { ... }

  func IsEntry(name string, entries []string) bool { ... }

  func CanManage(name string) bool { ... }

  func AllowGitHook() bool { ... }

  ```

#### Order of arguments of functions or methods should generally apply following rules (from left to right):

1. More important to less important.

2. Simpler types to more complicated types.

3. Same types should be put together whenever possible.

### Constant
- All constants MUST be written in UPPERCASE (SCREAMING_SNAKE_CASE)

- Words are separated by underscore `_`

- This rule applies to ALL constants, both internal and exported

```go

const MAX_RETRY = 3

const DEFAULT_TIMEOUT = 30

```

### Variables

- A variable name should follow general English expression or shorthand.

- Use camelCase

- In relatively simple (less objects and more specific) context, variable name can use simplified form as follows:

  - `userID` to `uid`

  - `repository` to `repo`

- If variable type is `bool`, its name should start with `Has`, `Is`, `Can` or `Allow`, etc.

  ```go

  var isExist bool

  var hasConflict bool

  var canManage bool

  var allowGitHook bool

  ```

### Other notes

- When something is waiting to be done, use comment starts with `TODO:` to remind maintainers.

- When a known problem/issue/bug needs to be fixed/improved, use comment starts with `FIXME:` to remind maintainers.

- When something is too magic and needs to be explained, use comment starts with `NOTE:`:

  ```Go

  // NOTE: os.Chmod and os.Chtimes don't recognize symbolic link,

  // which will lead "no such file or directory" error.

  return os.Symlink(target, dest)

  ```

- When dealing with security-related logic, use comment starts with `SECURITY:`

- When something important but is easy to overlook, use comment starts with `WARNING:`

- When the object is deprecated, use comment starts with `DEPRECATED:`:

  ```go

  // Email returns the user's oldest email, if one exists.

  // Deprecated: use Emails instead.

  func (r *UserResolver) Email(ctx context.Context) (string, error) {
    ...
  }

  ```