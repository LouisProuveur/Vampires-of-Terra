# Go Testing Crash Course

Oct 29, 2026 · @Louis

Go ships testing in the standard library (`testing` + `go test`), so you need no framework to write good tests. This course covers the toolchain, the idioms the community has converged on, and the pitfalls, with runnable examples. Code targets recent Go (1.24+); version-specific features are flagged.

## 1. Fundamentals

A test is a function named `TestXxx(t *testing.T)` in a file ending in `_test.go`; `go test` finds, compiles and runs it. Files ending in `_test.go` are excluded from normal builds, so test code never ships in your binary.

```go
// calc.go
package calc

func Add(a, b int) int { return a + b }
```

```go
// calc_test.go
package calc

import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("Add(2, 3) = %d, want %d", got, want)
    }
}
```

Run it:

```bash
go test            # current package
go test ./...      # every package under the current directory
go test -v ./...   # verbose: print each test and its result
go test -run TestAdd ./calc   # only tests matching the regexp
```

**Rules the toolchain enforces**

- The file name ends in `_test.go`; the function name starts with `Test` followed by an uppercase letter or digit (`TestAdd`, not `Testadd`).
- The signature is exactly `func(t *testing.T)`; there is no return value.
- Directories named `testdata`, or starting with `_` or `.`, are ignored by `./...`. Put fixtures in `testdata/`.

**Reporting failures**

| Method | Marks test failed | Stops the test | Use when |
| --- | --- | --- | --- |
| `t.Error` / `t.Errorf` | yes | no | Other checks are still meaningful; you see all failures in one run |
| `t.Fatal` / `t.Fatalf` | yes | yes | Continuing would panic or be meaningless (setup failed, nil result) |
| `t.Log` / `t.Logf` | no | no | Debug output; shown only on failure or with `-v` |
| `t.Skip` / `t.Skipf` | no | yes | Test cannot run here (missing env, `-short`) |

**Failure message convention:** name the function and its inputs, then `got` before `want`: `Add(2, 3) = 6, want 5`. A failing test should tell the reader what went wrong without opening the source.

Caveat: call `t.Fatal` / `t.FailNow` only from the goroutine running the test. From another goroutine, use `t.Error` and signal back through a channel.

## 2. Table-driven tests and subtests

The dominant Go idiom: describe cases as data, loop over them, and run each as a named subtest with `t.Run`. Adding a case costs one line.

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {name: "simple", a: 10, b: 2, want: 5},
        {name: "fraction", a: 1, b: 4, want: 0.25},
        {name: "negative", a: -9, b: 3, want: -3},
        {name: "divide by zero", a: 1, b: 0, wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Fatalf("Divide(%v, %v) error = %v, wantErr %v", tt.a, tt.b, err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**Why subtests matter**

- Each case reports separately: `--- FAIL: TestDivide/divide_by_zero`.
- You can run one case: `go test -run 'TestDivide/divide_by_zero'` (spaces become underscores; `/` separates levels).
- Each subtest can call `t.Parallel()`, `t.Skip()` and `t.Cleanup()` on its own.

**Tips**

- Give every case a descriptive `name`; it is your failure message.
- Use a `map[string]struct{...}` when you want unordered cases keyed by name; Go randomizes map iteration, which flushes out order dependencies.
- Since Go 1.22 the loop variable `tt` is per-iteration, so capturing it in a closure or `t.Parallel()` subtest is safe. Before 1.22 you needed `tt := tt`.
- Test errors by identity, not by string: `errors.Is(err, ErrDivideByZero)` or `errors.As`, rather than comparing `err.Error()`.
- If cases need very different setup or assertions, split them into separate tests; a table with a dozen `if tt.kind == ...` branches is a smell.

## 3. Helpers and test lifecycle

**Helpers.** Extract repeated setup into functions that take `*testing.T`, and call `t.Helper()` first so failures are reported at the caller's line, not inside the helper.

```go
func newTestStore(t *testing.T) *Store {
    t.Helper()
    s, err := NewStore(t.TempDir())
    if err != nil {
        t.Fatalf("NewStore: %v", err)
    }
    t.Cleanup(func() { s.Close() })
    return s
}
```

Convention: helpers that can fail take `t` and fail the test themselves (`t.Fatalf`) instead of returning an error the caller must check. Put `t` first (or `ctx` first, then `t`).

**Built-in lifecycle tools**

| API | What it does |
| --- | --- |
| `t.Cleanup(fn)` | Runs `fn` when the test (and its subtests) finish, in LIFO order. Prefer it to `defer` in helpers, because a `defer` in a helper runs when the helper returns |
| `t.TempDir()` | A fresh directory, deleted automatically afterwards |
| `t.Setenv(k, v)` | Sets an env var for this test only and restores it; cannot be used with `t.Parallel()` |
| `t.Chdir(dir)` | Changes working directory for the test and restores it (Go 1.24+) |
| `t.Context()` | A `context.Context` cancelled just before cleanups run (Go 1.24+) |
| `t.Skip(...)` | Skips with a reason, e.g. on the wrong OS or when a service is missing |
| `t.Name()` | The full test name, handy for unique resource names |

```go
func TestConfigFromEnv(t *testing.T) {
    t.Setenv("APP_PORT", "9090")

    cfg, err := LoadConfig()
    if err != nil {
        t.Fatal(err)
    }
    if cfg.Port != 9090 {
        t.Errorf("Port = %d, want 9090", cfg.Port)
    }
}
```

Prefer these to hand-rolled `os.Setenv` / `os.MkdirTemp` / `defer` code: they clean up even when the test fails or panics, and they respect subtests.

## 4. Assertions: stdlib, go-cmp, testify

Go has no built-in `assert`. The official stance (Go FAQ and Code Review Comments) is that plain `if` + `t.Errorf` gives clearer failures and keeps tests as ordinary Go. In practice, teams pick one of three styles.

| Style | Strengths | Trade-offs |
| --- | --- | --- |
| Plain `if` + `t.Errorf` | Zero dependencies, full control of the message | Verbose for structs, slices, maps |
| `github.com/google/go-cmp/cmp` | Precise, readable diffs for deep comparisons; options for ignoring fields | Still hand-written `if` around it |
| `github.com/stretchr/testify` (`assert`, `require`) | Concise; `require` stops on failure; also offers `mock` and `suite` | Extra dependency; messages are less tailored; `suite` and `mock` are often avoided |

**Deep comparison with go-cmp** (prefer it over `reflect.DeepEqual`, which gives no diff and trips on unexported fields):

```go
import (
    "testing"
    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseUser(t *testing.T) {
    got, err := ParseUser(`{"name":"Ada","tags":["x","y"]}`)
    if err != nil {
        t.Fatal(err)
    }
    want := User{Name: "Ada", Tags: []string{"x", "y"}}

    if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(User{}, "CreatedAt")); diff != "" {
        t.Errorf("ParseUser() mismatch (-want +got):\n%s", diff)
    }
}
```

**The same with testify**:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParseUser(t *testing.T) {
    got, err := ParseUser(`{"name":"Ada"}`)
    require.NoError(t, err)          // require = Fatal: stop here on failure
    assert.Equal(t, "Ada", got.Name) // assert = Error: keep going
    assert.ErrorIs(t, Validate(User{}), ErrEmptyName)
}
```

Note the argument order: testify takes `(expected, actual)`, while the Go convention in messages is `got` then `want`. Mixing them up produces confusing output.

**Guidance**

- Follow the conventions of the codebase you are in; consistency beats purity.
- For floats, compare with a tolerance (`math.Abs(got-want) > 1e-9`, or `cmpopts.EquateApprox`).
- Compare errors with `errors.Is` / `errors.As`, not strings.

## 5. Package layout, testdata and golden files

**Internal vs external test packages.** Test files may declare either the package under test or its `_test` twin:

```go
package store        // internal (white-box): can touch unexported identifiers
package store_test   // external (black-box): imports store like any consumer
```

- Default to **external** (`store_test`) for tests of the public API. It proves your API is usable and avoids import cycles when the test needs a helper package that itself imports yours.
- Use **internal** tests when you must exercise unexported logic directly.
- Both can live in the same directory. A common bridge is `export_test.go` (package `store`) that re-exports an internal symbol for the external tests only; it is compiled only during tests.

**Typical layout**

```text
myapp/
  store/
    store.go
    store_test.go          # package store_test
    export_test.go         # package store: exposes internals to tests
    testdata/
      users.json           # fixtures; ignored by ./...
      parse.golden         # expected output
```

Tests run with the package directory as the working directory, so `os.ReadFile("testdata/users.json")` just works.

**Golden files** compare output against a stored file, which suits large or structured output (rendered HTML, generated code, JSON). A `-update` flag regenerates them:

```go
var update = flag.Bool("update", false, "rewrite golden files")

func TestRender(t *testing.T) {
    got := Render(sampleInput)
    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {
        if err := os.WriteFile(golden, got, 0o644); err != nil {
            t.Fatal(err)
        }
    }
    want, err := os.ReadFile(golden)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(got, want) {
        t.Errorf("output differs from %s; run with -update to accept", golden)
    }
}
```

```bash
go test ./store -update   # regenerate, then review the git diff before committing
```

Always review a golden-file diff in code review: `-update` makes it easy to bless a bug.

## 6. Isolation: interfaces, fakes, stubs and mocks

Go's implicit interfaces make testing easy: depend on a small interface, and substitute a test implementation. The proverb: *accept interfaces, return structs*, and define the interface **where it is consumed**, with only the methods you call.

```go
// production code
type UserStore interface {
    GetUser(ctx context.Context, id string) (User, error)
}

type Greeter struct{ store UserStore }

func NewGreeter(s UserStore) *Greeter { return &Greeter{store: s} }

func (g *Greeter) Greet(ctx context.Context, id string) (string, error) {
    u, err := g.store.GetUser(ctx, id)
    if err != nil {
        return "", fmt.Errorf("greet %s: %w", id, err)
    }
    return "Hello, " + u.Name, nil
}
```

**Three kinds of test doubles**

| Double | What it is | Use for |
| --- | --- | --- |
| Stub | Returns canned values | Driving a code path (error, empty result) |
| Fake | Working lightweight implementation (in-memory map, `fstest.MapFS`) | Behavior that depends on state |
| Mock / spy | Records calls and asserts on them | Verifying an interaction that is the point (an email was sent once) |

A hand-written fake is usually the best choice: it is plain Go, no codegen, and survives refactoring.

```go
type fakeStore struct {
    users map[string]User
    calls int
}

func (f *fakeStore) GetUser(_ context.Context, id string) (User, error) {
    f.calls++
    u, ok := f.users[id]
    if !ok {
        return User{}, ErrNotFound
    }
    return u, nil
}

func TestGreet(t *testing.T) {
    g := NewGreeter(&fakeStore{users: map[string]User{"1": {Name: "Ada"}}})

    got, err := g.Greet(t.Context(), "1")
    if err != nil {
        t.Fatal(err)
    }
    if want := "Hello, Ada"; got != want {
        t.Errorf("Greet = %q, want %q", got, want)
    }

    _, err = g.Greet(t.Context(), "missing")
    if !errors.Is(err, ErrNotFound) {
        t.Errorf("err = %v, want ErrNotFound", err)
    }
}
```

**Other isolation techniques**

- **Function fields / injected funcs** for tiny seams: `type Service struct{ now func() time.Time }`. Inject a fixed clock instead of calling `time.Now()` directly, and `rand.Rand` with a fixed seed.
- **`testing/fstest.MapFS`** to fake a filesystem when your code takes an `fs.FS`.
- **Generated mocks** (`go.uber.org/mock` / `mockgen`, `mockery`) pay off for large interfaces or strict call-order checks; they cost more maintenance and tie tests to implementation details.
- Avoid mocking types you do not own (a database driver, `http.Client`); wrap them behind your own small interface, or use a real fake like `httptest` (next section).

Rule of thumb: test behavior and outcomes, not which internal methods got called.

## 7. HTTP testing with httptest

`net/http/httptest` offers two tools. **`NewRecorder`** tests a handler in-process, with no network. **`NewServer`** starts a real local server, which is what you want to test HTTP *clients*.

**Testing a `net/http` handler**

```go
func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rec := httptest.NewRecorder()

    HealthHandler(rec, req)

    res := rec.Result()
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
    }
    body, _ := io.ReadAll(res.Body)
    if got := strings.TrimSpace(string(body)); got != `{"status":"ok"}` {
        t.Errorf("body = %s", got)
    }
}
```

**Testing a Gin router.** Build the router in a function (`func NewRouter(deps) *gin.Engine`) so tests can construct it, then call `ServeHTTP` directly:

```go
func TestCreatePlayer(t *testing.T) {
    gin.SetMode(gin.TestMode) // silences debug logs
    router := NewRouter(&fakeStore{})

    body := strings.NewReader(`{"name":"Ada"}`)
    req := httptest.NewRequest(http.MethodPost, "/players", body)
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    if rec.Code != http.StatusCreated {
        t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body)
    }
    var got Player
    if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
        t.Fatal(err)
    }
    if got.Name != "Ada" {
        t.Errorf("Name = %q, want Ada", got.Name)
    }
}
```

The same pattern works for chi, echo and the standard `http.ServeMux`, since all implement `http.Handler`.

**Testing an HTTP client against a fake server**

```go
func TestFetchQuote(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/quote" {
            http.NotFound(w, r)
            return
        }
        fmt.Fprint(w, `{"text":"hello"}`)
    }))
    t.Cleanup(srv.Close)

    c := NewClient(srv.URL) // base URL is injectable
    q, err := c.FetchQuote(t.Context())
    if err != nil {
        t.Fatal(err)
    }
    if q.Text != "hello" {
        t.Errorf("Text = %q", q.Text)
    }
}
```

Other handy members: `httptest.NewTLSServer` for HTTPS, and `srv.Client()` for a client pre-configured to trust it. For WebSocket code, the same `httptest.NewServer` plus a dialer against `ws://` + `srv.Listener.Addr()` gives a real end-to-end test.

Tip: design code so the base URL, `http.Client` and clock are injected; that single decision removes most need for mocks.

## 8. Parallel tests and concurrency

**`t.Parallel()`** lets a test run alongside other parallel tests in the same package (different packages already run in parallel by default). Call it first, before any other work:

```go
func TestSlugify(t *testing.T) {
    tests := []struct{ name, in, want string }{
        {"spaces", "Hello World", "hello-world"},
        {"symbols", "Go!", "go"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // tt is per-iteration since Go 1.22
            if got := Slugify(tt.in); got != tt.want {
                t.Errorf("Slugify(%q) = %q, want %q", tt.in, got, tt.want)
            }
        })
    }
}
```

- Parallel tests pause until the sequential tests of the package finish, then run up to `-parallel N` at once (default `GOMAXPROCS`).
- Only parallelize tests that share no mutable state: no globals, no fixed ports, no shared files. Use `t.TempDir()` and port `0` (`httptest` and `net.Listen("tcp", ":0")` pick a free one).
- `t.Setenv` and `t.Chdir` panic in parallel tests, because they change process-wide state.

**Race detector.** Run it routinely in CI; it catches data races at runtime, but only on code paths your tests actually execute:

```bash
go test -race ./...
```

**Testing goroutines without `time.Sleep`.** Sleeping makes tests slow and flaky. Instead, synchronize explicitly: channels, `sync.WaitGroup`, or `context` with a timeout.

```go
func TestWorkerProcesses(t *testing.T) {
    ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
    defer cancel()

    results := make(chan int, 1)
    go Worker(ctx, 21, results)

    select {
    case got := <-results:
        if got != 42 {
            t.Errorf("got %d, want 42", got)
        }
    case <-ctx.Done():
        t.Fatal("timed out waiting for worker")
    }
}
```

**Testing time-dependent code.** Go 1.25 added `testing/synctest`, which runs code in a "bubble" with a fake clock that advances instantly once all goroutines are blocked, so a test of a 10-minute timeout runs in microseconds:

```go
import "testing/synctest"

func TestCacheExpires(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        c := NewCache(time.Minute)
        c.Set("k", "v")

        time.Sleep(2 * time.Minute) // virtual time: returns immediately
        synctest.Wait()             // let background goroutines settle

        if _, ok := c.Get("k"); ok {
            t.Error("entry should have expired")
        }
    })
}
```

On older Go versions, inject a clock (`now func() time.Time`) instead.

**Goroutine leaks.** `go.uber.org/goleak` fails a test, or a whole package via `TestMain`, if goroutines are still running afterwards:

```go
func TestMain(m *testing.M) { goleak.VerifyTestMain(m) }
```

## 9. Integration tests, TestMain, build tags and -short

Tests that need a database, a network service or minutes of runtime should be easy to exclude from the fast inner loop. Go offers three mechanisms.

**1. `testing.Short()`** with the `-short` flag: a convention for "skip the slow stuff".

```go
func TestEndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in -short mode")
    }
    // ...
}
```

```bash
go test -short ./...   # fast
go test ./...          # everything
```

**2. Build tags**: the file is not even compiled unless you ask for it. The `//go:build` line must be at the top, before `package`, followed by a blank line.

```go
//go:build integration

package store_test

func TestPostgresStore(t *testing.T) { /* needs a real DB */ }
```

```bash
go test -tags=integration ./...
```

**3. `TestMain`**: one per package, for expensive setup and teardown that wraps every test (start a container, migrate a DB, seed fixtures). Call `m.Run()` and handle its exit code; since Go 1.15, simply returning from `TestMain` exits with that result, so `defer` works if you do not call `os.Exit`.

```go
var testDB *sql.DB

func TestMain(m *testing.M) {
    var err error
    testDB, err = sql.Open("postgres", os.Getenv("TEST_DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    if err := migrate(testDB); err != nil {
        log.Fatal(err)
    }

    code := m.Run()

    testDB.Close()
    os.Exit(code) // optional since Go 1.15; os.Exit skips defers, so clean up above
}
```

**Practical guidance**

- Keep `TestMain` for process-wide setup. Prefer per-test helpers with `t.Cleanup` for anything a single test owns.
- For real dependencies (Postgres, Redis, Kafka), `testcontainers-go` starts throwaway Docker containers from your test code. Give each test its own schema, transaction (rolled back at the end), or uniquely named resources so tests stay independent and can run in parallel.
- Run integration tests in CI as a separate job so a flaky service does not block the unit-test signal.
- Name them clearly (`TestIntegration_...`) so `-run Integration` works without tags if you prefer that approach.

## 10. Benchmarks, fuzzing and examples

The `testing` package covers three more kinds of function, all in `_test.go` files.

### Benchmarks

Named `BenchmarkXxx(b *testing.B)`. Since Go 1.24, `b.Loop()` is the preferred form: it handles timing and keeps the compiler from optimizing the work away. (Older code uses `for i := 0; i < b.N; i++`.)

```go
func BenchmarkSlugify(b *testing.B) {
    for b.Loop() {
        Slugify("Hello, World! Again")
    }
}
```

```bash
go test -bench=. -benchmem -run='^$' ./...   # -run='^$' skips regular tests
go test -bench=Slugify -count=10 > old.txt   # collect several runs
# change code, collect new.txt, then compare statistically:
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

Use `-benchmem` for allocations per op, and `b.ReportAllocs()` inside a benchmark to always report them. Profile with `-cpuprofile cpu.out` or `-memprofile mem.out`, then `go tool pprof`. Never trust a single run; use `benchstat` over several.

### Fuzzing (Go 1.18+)

The fuzzer mutates inputs to find ones that break an invariant. Good for parsers, decoders and anything handling untrusted input.

```go
func FuzzReverse(f *testing.F) {
    f.Add("hello") // seed corpus
    f.Add("")
    f.Fuzz(func(t *testing.T, s string) {
        got := Reverse(Reverse(s))
        if got != s {
            t.Errorf("Reverse(Reverse(%q)) = %q", s, got)
        }
    })
}
```

```bash
go test                       # runs seeds + saved corpus as ordinary tests
go test -fuzz=FuzzReverse -fuzztime=30s
```

A failing input is saved under `testdata/fuzz/FuzzReverse/`; commit it, and it becomes a permanent regression test. Fuzz functions assert **properties** (round trips, no panics, output is valid), since there is no expected value.

### Example tests

`ExampleXxx()` functions are compiled, run, and shown in `go doc` / pkg.go.dev. If the function ends with an `// Output:` comment, `go test` verifies it, so documentation can never go stale.

```go
func ExampleSlugify() {
    fmt.Println(Slugify("Hello World"))
    // Output: hello-world
}
```

Naming ties them to what they document: `ExampleSlugify`, `ExampleStore_Get` (method), `ExampleSlugify_withSymbols` (suffix, lowercase first letter).

## 11. Coverage, flags and command cheatsheet

**Coverage**

```bash
go test -cover ./...                          # percentage per package
go test -coverprofile=cover.out ./...         # write a profile
go tool cover -func=cover.out                 # per-function summary
go tool cover -html=cover.out                 # browse annotated source
go test -coverpkg=./... -coverprofile=c.out ./...   # count coverage across packages
```

Since Go 1.20 you can also collect coverage from a running binary built with `go build -cover`, useful for end-to-end tests. Treat coverage as a map of untested code, not a target: 100% can still miss every important edge case, and chasing the number invites meaningless tests.

**Flags you will use daily**

| Flag | Effect |
| --- | --- |
| `-v` | Verbose: list each test and its `t.Log` output |
| `-run regexp` | Run only matching tests; use `A/B` to target subtests |
| `-skip regexp` | Skip matching tests (Go 1.20+) |
| `-count=N` | Run N times; `-count=1` disables result caching |
| `-race` | Enable the race detector |
| `-short` | Set `testing.Short()` to true |
| `-timeout 30s` | Fail if the package takes longer (default 10m) |
| `-failfast` | Stop after the first failing test |
| `-shuffle=on` | Randomize test order to expose inter-test dependencies |
| `-parallel N` | Max parallel tests per package |
| `-json` | Machine-readable output for CI and tooling |
| `-tags x` | Enable build tag `x` |
| `-bench`, `-benchmem`, `-fuzz` | Benchmarks and fuzzing (section 10) |

**Test caching.** `go test` caches successful results when the code, test files, flags and any files or env vars the test read are unchanged; the output shows `(cached)`. Use `-count=1` to force a rerun, for example for tests that hit external services.

**Package arguments:** `go test` (current dir), `go test ./...` (recursive), `go test ./store ./api` (specific), `go test .` (same as no argument but explicit).

**Static checks that run with go test.** `go test` runs a small, high-confidence subset of `go vet` automatically (printf formats, copied locks, and so on). Run `go vet ./...` and a linter such as `golangci-lint` in CI as well; `testifylint` and `paralleltest` are lint rules aimed at test code.

**Typical CI line**

```bash
go test -race -shuffle=on -coverprofile=cover.out ./...
```

## 12. Conventions, best practices and anti-patterns

**Do**

- Test **behavior through the public API**; unexported details change, contracts should not.
- Write failures that explain themselves: `FuncName(input) = <got>, want <want>`.
- Keep tests **independent and order-free**: no shared mutable globals, no reliance on another test having run. Verify with `-shuffle=on`.
- Make code testable by design: small interfaces, injected dependencies (clock, HTTP client, store, logger), constructors that take config instead of reading globals.
- Use `t.Helper()`, `t.Cleanup`, `t.TempDir`, `t.Setenv` and `t.Context()` rather than manual equivalents.
- Cover the unhappy paths: errors, empty input, boundaries, nil, concurrency. Add a regression test for every bug before fixing it.
- Run `go test -race ./...` in CI; keep the default suite fast (seconds), and gate slow tests behind `-short` or build tags.
- Make tests deterministic: seed random sources, inject clocks, sort map output before comparing.
- Name tests `TestType_Method` or `TestFunction_scenario`; subtests describe the case in plain words.

**Don't**

| Anti-pattern | Why it hurts | Instead |
| --- | --- | --- |
| `time.Sleep` to wait for goroutines | Slow and flaky | Channels, `WaitGroup`, `synctest`, polling with a deadline |
| Asserting on full error strings | Breaks on rewording | `errors.Is` / `errors.As`, or check a stable substring only when unavoidable |
| Tests that hit the real internet or a shared DB | Flaky, slow, order-dependent | `httptest`, fakes, isolated containers |
| Logic in tests (loops, conditionals, reimplementing the code) | The test can contain the same bug | Literal expected values; keep the table-driven loop trivial |
| Over-mocking | Tests pin implementation, break on refactors | Fakes; assert outcomes |
| `t.Fatal` in goroutines | Undefined behavior: only the test goroutine may call it | `t.Error` and a channel |
| Ignoring `-race` findings | Real bugs in production | Fix them; never merely retry |
| Chasing 100% coverage | Brittle, low-value tests | Cover risk, not lines |
| Global state mutated by tests (`os.Setenv`, package vars) | Cross-test pollution | `t.Setenv`, injected config, restore in `t.Cleanup` |

**Handy standard-library helpers:** `testing/iotest` (readers that fail or return one byte at a time), `testing/fstest` (in-memory `fs.FS`), `testing/quick` (property-based checks, largely superseded by fuzzing), `net/http/httptest`, and `slices.Sorted(maps.Keys(m))` to make map output deterministic.

**Further reading:** the Go blog post "Using Subtests and Sub-benchmarks", the Go wiki page "TableDrivenTests", the official [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) (test sections), the `testing` package docs at pkg.go.dev, and the fuzzing tutorial at go.dev/doc/tutorial/fuzz.
