### Review priority

1. Correctness, safety, and alignment with **dp1-go** / DP-1 (no ad-hoc canonicalization).
2. Concurrency and cancellation correctness on I/O paths (`context`, HTTP client, fetch).
3. Architecture and package-boundary discipline (`cmd` vs `internal/*`).
4. **CLI contract** clarity: flags, env precedence, `--json` shapes, stable error reporting (`internal/output`).
5. Test and documentation sufficiency (`docs/cli_design.md`, `docs/architecture.md` when boundaries move).

### Required expanded review posture

- Do not review only for local diff correctness.
- Infer the intended operator outcome (CLI UX and scripts consuming `--json`), then judge whether the change delivers it safely.
- Prefer findings about behavior, correctness, maintenance risk, DP-1 interop, or weak contracts over style-only comments.
- Do not speculate. Only raise issues that are concrete and actionable.

### dp1-cli-specific review focus

- **No duplicated protocol logic** in `cmd`: validation and signing must go through dp1-go public APIs.
- **User-visible changes** stay documented in `docs/cli_design.md`; feed paths and 201 semantics for `publish` remain consistent with stated behavior.
- **Credentials:** resolution order (flags, env, config) is understandable and error messages are actionable.
- **Output:** human vs `--json` paths stay consistent; secrets are never printed.

### Go-specific review focus

- Error handling is explicit, contextual, and does not hide root causes.
- Package boundaries avoid circular or muddy dependencies.
- Concurrency has clear ownership when present; contexts are passed through network calls.
- Comments explain non-obvious intent where the comment contract applies.

### Hindsight and refactor review

After reading the implementation, consider whether the goal would be better served by:

- deleting complexity
- simplifying a package boundary
- narrowing a flag surface
- moving HTTP or FS details behind a smaller interface

Only include findings when there is a clearly better alternative.

### Tests and docs sufficiency

Assess only real gaps:

1. Do unit tests cover core logic and edge cases?
2. Do tests cover boundary behavior (e.g. feed error mapping, config edge cases) where it matters?
3. Are failure paths meaningfully exercised?
4. Should `docs/cli_design.md` or `docs/architecture.md` be updated?

### Preferred review output shape

Use only sections that have real content:

1. Critical correctness issues
2. Concurrency or lifecycle issues
3. Architecture or CLI contract issues
4. Better alternative designs
5. Test gaps
6. Documentation gaps

If there are no meaningful findings, keep the review brief.

### Verdict

End your review with exactly one line:

- `Verdict: accept`
- `Verdict: revise`
