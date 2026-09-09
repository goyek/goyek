# Codex Security findings verification

Verified on 2026-07-21 against `main` at
`b290996180dc057ad1b6c131f832737811ff71b5`. The local branch was clean and
matched a freshly fetched `origin/main` (`0` ahead, `0` behind).

## Outcome

- 12 goyek findings were reviewed. The initial scrape returned 11; a continuous
  scan published `299cdeb...` during verification, so it was reviewed too.
- 5 findings remain open because the reported condition is still present and is
  security-relevant or worthwhile CI hardening.
- 7 findings were closed as `false_positive` in Codex Security.
  - 6 describe real, current correctness or robustness defects, but require
    trusted in-process code and do not cross a security boundary.
  - 1 (`299cdeb...`) reports behavior that is explicitly required by the
    documented low-level API contract.
- No finding was marked `fixed`: none of the reported code conditions has been
  removed from `main`.

The supplied URL has two different scopes: its `repo` query selects goyek, but
its path selection `35685b7d8ad48191b113003a8826d9b6` resolves to an
`opentelemetry-python-contrib` finding. That unrelated finding was not changed.
All decisions below use the goyek repository filter.

## Verification method

- Playwright 1.61.1 was used to retrieve the authenticated finding list,
  individual finding details, scan state, and post-update status lists from the
  same backend used by the Codex Security page.
- Both active statuses (`new`, `triaged`, `in_progress`) and archived statuses
  (`fixed`, `wontfix`, `duplicate`, `false_positive`) were queried.
- The final status reconciliation returned exactly 5 active and 7 archived
  findings. The continuous scanner still listed current commit `b290996...` as
  pending and its last automated scan was `d697ab9...`; `b290996...` changes
  only `actions/checkout` pins. Every finding was nevertheless inspected
  manually against `b290996...`, including the checkout-specific evidence.
- Every claim was checked against current files and relevant history. Focused
  reproductions covered CLI task selection, tag-name shell substitution,
  middleware default-task visibility, stale task handles, dependency-slice
  aliasing, nil contexts, and derived-context cleanup ordering.
- `go test -race ./...`, `go vet ./...`, and the build-module tests/vet passed.

## Findings kept open

### 1. `19c6c53d98b481919e82e70af0abc8b7`

**CI lint task runs mutable Docker image on writable checkout**  
Codex severity: high. Recommended triage: medium. **Valid; keep open.**

`build/mdlint.go:25-28` still runs
`ghcr.io/igorshubovych/markdownlint-cli:v0.41.0` by mutable tag, mounts the whole
checkout at `/workdir` without `:ro`, and leaves container networking enabled.
The task is part of `lint` (`build/lint.go:8-12`), and tests run later from the
same writable checkout (`build/test.go:9-17`). A compromised image can therefore
rewrite Go or test files that a later host process executes.

One part of the old evidence is stale: the workflow now uses checkout v7, which
inherits checkout v6's move of persisted credentials from `.git/config` to
`$RUNNER_TEMP`. The container does not directly receive that directory. This
reduces direct credential exposure, but it does not remove the writable
checkout-to-host-execution chain. The workflow also now declares
`permissions: contents: read`.

### 2. `b4f9a26369fc819190d7d25f6f757f24`

**Default task hidden from executor middlewares**  
Codex severity: medium. Recommended triage: medium for affected embedders.
**Valid; keep open.**

`Flow.Execute` installs executor middleware and passes the caller's original
task slice at `flow.go:390-403`. Only the inner executor expands an empty slice
to the default task at `executor.go:54-58`. An authorization or audit middleware
can therefore observe no task while the default task actually runs. goyek does
not itself expose a remote authorization boundary, so impact is conditional on
an embedder using `UseExecutor` for policy enforcement.

### 3. `78bc64ce67b8819196e0530d4e4da9d4`

**GitHub Actions tag name command injection**  
Codex severity: medium. Recommended triage: low. **Valid primitive; keep open
for hardening.**

`.github/workflows/release.yml:3-5` accepts `v*` tags, and line 18 inserts
`${{ github.ref_name }}` directly into an unquoted shell command. Git accepts
tag names containing command substitutions and shell operators; a safe local
reproduction confirmed substitution. Exploitation requires tag-push rights,
while the job has no checkout or explicit secrets and only `contents: read`, so
there is no demonstrated privilege gain. The interpolation should nevertheless
be removed; GitHub's guidance recommends passing potentially unsafe context
values through an intermediate environment variable rather than inserting them
directly into scripts.

### 4. `9cdee0f6fccc8191bb52cf8f2b725e1b`

**Unbounded task log buffering can exhaust memory**  
Codex severity: medium. Recommended triage: low availability/robustness.
**Valid but narrower than originally reported; keep open.**

The original core `taskflow.go`/`tf.go` path no longer exists, and
`runner.go:87-90` streams to the supplied output. Equivalent unbounded buffers
remain in `middleware/verbose.go:13-23` and
`middleware/bufferparallel.go:12-24`. `SilentNonFailed` is enabled by the build
when `-v` is absent (`build/main.go:61-63`); current CI passes `-v`, so the main
CI job is not on that particular path. Direct users can still exhaust memory
when attacker-influenced or simply very large task output is buffered.

### 5. `2297c861123c81918d4d52f7360c2aed`

**CI task selection regression skips diff gate**  
Codex severity: low. Recommended triage: low CI integrity. **Valid; keep open.**

The workflow invokes `go run . -v ci` at
`.github/workflows/build.yml:21`. `SplitTasks` recognizes tasks only before the
first flag (`split.go:26-41`), while `build/main.go:45-49,78` ignores the
remaining parsed positional argument. Execution therefore falls back to
default task `all`, which omits `diff`; task `ci` includes it.

Focused dry runs confirmed the difference:

- `go run . -dry-run -v ci` ran through `all` only.
- `go run . ci -dry-run -v` also ran `diff` and `ci`.

## Findings closed in Codex Security

All seven entries below are now archived with status `false_positive`. The six
correctness defects remain candidates for the ordinary engineering backlog;
closing them here means only that they are not security vulnerabilities in the
repository's threat model.

| Codex ID | Finding | Closure reason |
| --- | --- | --- |
| `d8bab7728a848191b8070ec409b0e757` | DefinedTask refactor exposes mutable graph invariants | Real API-integrity bug, but only trusted in-process code holding task objects or dependency slices can trigger it. Such code already controls task actions. |
| `898a32e58f888191a457ea9ecfbe6280` | Derived A contexts are not canceled before cleanup | Real cleanup-ordering hang, but the trigger is trusted task code that can already block the build process directly. |
| `b3f4442eb3a08191829fa63ffd061d46` | Nil context passed to Flow.Execute now panics | Real compatibility/reliability regression; the input is an in-process Go context, not an attacker-controlled CLI or network value. |
| `9936537c809c81918ae7b4aef03d33ef` | ReportLongRun can crash on non-positive duration | Real exported-API robustness bug. The checked-in CLI guards `d > 0`, and only trusted configuration can reach the crash. |
| `d5a931446b588191ae930896ea5fa69b` | Stale task handles can undefine replacement tasks | Real registry-integrity bug; exploiting it requires trusted same-process code that can already define and undefine tasks. |
| `9865cadff6448191b71332562a90d28d` | TF Log concatenates arguments without spaces | The issue survives under `A`/`CodeLineLogger`, but malformed human-readable output is a correctness defect with no executable or authorization sink. |
| `299cdebfd07881919411ea84f82b7a8a` | Direct NewRunner output synchronization regression | Not a defect: `Input.Output` explicitly requires a concurrency-safe non-nil writer, `NewRunner` explicitly preserves it, and callers are directed to `SyncWriter`. The reported reproduction violates that contract. |

## Remediation plan

### Phase 1: CI and release boundaries

1. **Remove release tag interpolation (`78bc64ce...`).**
   - Put `github.ref_name` and `github.repository` in step-level `env` values.
   - Validate the tag against the project's accepted semantic-version format.
   - Quote every shell expansion and use `curl --fail --show-error`.
   - Add an actionlint check and a test showing values such as `v$(id)` remain
     literal data.
2. **Harden markdownlint execution (`19c6c53d...`).**
   - Pin a reviewed image to its immutable digest, not only a version tag.
   - Mount only required markdown/config paths read-only; avoid mounting `.git`.
   - Add `--network none`, a non-root user, and dropped capabilities where the
     image supports them.
   - Set checkout `persist-credentials: false` because later steps do not need
     authenticated Git operations.
   - Prefer a pinned native/tool dependency over Docker if maintenance cost is
     acceptable.
   - Test the exact Docker argument vector and task ordering.
3. **Restore the CI diff gate (`2297c861...`).**
   - Change the workflow to task-first syntax: `go run . ci -v`.
   - Make `build/main.go` reject unexpected `flag.Args()` so flag-first task
     input fails loudly instead of selecting the default.
   - Add a dry-run smoke test using the exact workflow command and assert that
     `diff` and `ci` both execute.

### Phase 2: Execution-policy semantics

4. **Expose the effective default task to middleware (`b4f9a263...`).**
   - In `Flow.Execute`, copy the caller's task slice and resolve the default
     before applying executor middleware.
   - Keep direct executor validation consistent and do not mutate the caller's
     backing array.
   - Add an authorization-middleware regression test: middleware must see the
     default task, denial must prevent its action, and audit output must name it.

### Phase 3: Bound output retention

5. **Replace unbounded middleware buffers (`9cdee0f6...`).**
   - Audit both `SilentNonFailed` and `BufferParallel` together.
   - Use a configurable cap with a clear truncation marker, or spill to a
     temporary file after a small in-memory threshold.
   - Preserve failure replay, successful-output discard, per-task grouping,
     concurrent-write safety, and cleanup on panic/cancellation.
   - Test just below/above the threshold and run the cases under `-race`.

### Phase 4: Ordinary correctness backlog

These items were closed only in the security tracker. Address them after the
open security/CI work, preferably as small focused changes:

1. Fix task identity and dependency ownership together (`d8bab772...`,
   `d5a93144...`): copy dependency slices, reject nil/cross-flow/stale handles,
   and require `f.tasks[name] == task` for APIs accepting a task handle.
2. Fix derived-context lifecycle (`898a32e5...`): keep derived cancel functions
   in shared lifecycle state and invoke all cancels before any user cleanup.
3. Normalize nil contexts before executor middleware and defensively in the
   executor (`b3f4442e...`).
4. Validate `ReportLongRun` duration synchronously, preferably treating
   non-positive values as disabled (`9936537c...`).
5. Restore Println-compatible formatting in `CodeLineLogger.Log` and add exact
   output tests (`9865cadf...`).
6. No code change is needed for `299cdebf...`; an extra documentation
   cross-reference is optional.

## Completion gates

For each remediation change:

- run `go test -race ./...` in the root module;
- run `go test ./...` and `go vet ./...` in `build`;
- run `go vet ./...` in the root module;
- run the focused new regression tests repeatedly under `-race` where
  concurrency is involved;
- run `git diff --check` and the corrected task-first CI dry run;
- close an open Codex finding as `fixed` only after its fixing PR is merged and
  record the PR URL in the resolution.

## References

- [actions/checkout v7 README at the pinned commit](https://github.com/actions/checkout/blob/3d3c42e5aac5ba805825da76410c181273ba90b1/README.md)
- [GitHub Actions script-injection guidance](https://docs.github.com/en/actions/concepts/security/script-injections)
