---
name: qa-debugging
description: >
  Systematic debugging and quality assurance workflow. Trigger: when investigating
  bugs, test failures, runtime errors, unexpected behavior, performance issues,
  or when the user says "debug", "fix this error", "why is this failing",
  "something is wrong", or provides an error message/stack trace.
---

# QA Debugging

Systematic methodology for diagnosing and fixing bugs. This skill replaces
ad-hoc "try things until it works" with a structured investigation process
that identifies root causes, not symptoms.

The core principle: understand the bug before writing the fix. A fix without
understanding is a patch that will break again.

## When to Activate

- User reports a bug or unexpected behavior
- Test suite fails
- Runtime error or stack trace is provided
- User says "debug", "fix", "broken", "failing", "error", "why is this..."
- Build or compilation fails
- Performance degradation is observed
- Flaky test investigation

## The Debugging Protocol

Follow this sequence. Do not skip steps.

```
1. REPRODUCE  → Can you trigger the bug reliably?
2. ISOLATE    → What is the minimal reproduction?
3. DIAGNOSE   → What is the root cause?
4. HYPOTHESIZE → What do you think is wrong and why?
5. VERIFY     → Prove the hypothesis with evidence
6. FIX        → Write the minimal correct fix
7. VALIDATE   → Confirm the fix resolves the issue without regressions
8. DOCUMENT   → Save the finding for future reference
```

### Step 1: REPRODUCE

Before investigating, confirm the bug exists and is reproducible.

```
# Run the failing test/command exactly as reported
# Capture the full error output
# Note: environment, OS, versions if relevant
```

If the bug is not reproducible:
- Ask for exact reproduction steps
- Check if it's environment-specific
- Check if it's timing-dependent (race condition)
- Check recent changes (git log, git diff)

### Step 2: ISOLATE

Narrow down the scope. The goal is the smallest possible reproduction.

Techniques:
- **Binary search**: comment out half the code, does it still fail?
- **Minimal test case**: write the simplest test that triggers the bug
- **Dependency elimination**: does it fail with mocked dependencies?
- **Git bisect**: `git bisect start`, `git bisect bad`, `git bisect good <commit>`

Output: "The bug is in [file:function] — triggered when [condition]"

### Step 3: DIAGNOSE

Read the code path that triggers the bug. Trace the execution flow.

Key questions:
- What are the inputs at the failure point?
- What is the expected state vs actual state?
- What changed recently? (`git log --oneline -20`, `git diff HEAD~5`)
- Are there related issues in the codebase? (grep for similar patterns)

Tools to use:
- Read the failing code and its callers
- Check test coverage for the area
- Search for similar patterns that might have the same bug
- Read error handling — is an error being swallowed?

### Step 4: HYPOTHESIZE

State your hypothesis clearly before attempting a fix:

```
HYPOTHESIS: [what you think is wrong]
EVIDENCE: [what supports this hypothesis]
PREDICTION: [if this hypothesis is correct, then X should be true]
```

This prevents "shotgun debugging" — changing random things hoping something works.

### Step 5: VERIFY

Test the hypothesis before writing the fix:

- Add a targeted assertion or log that would confirm/deny the hypothesis
- Run the reproduction case with the diagnostic in place
- If the hypothesis is wrong, go back to Step 3

Anti-pattern: skipping verification and going straight to "fix". This leads to
patches that mask the real bug.

### Step 6: FIX

Write the minimal correct fix. Not the biggest refactor. Not the cleanest code.
The smallest change that correctly addresses the root cause.

Rules:
- Fix the root cause, not the symptom
- One fix per bug — do not bundle unrelated changes
- If the fix requires a design change, flag it and propose separately
- Preserve existing behavior for all non-buggy cases

### Step 7: VALIDATE

Confirm the fix works and doesn't break anything:

1. Run the original failing test/reproduction → should pass
2. Run the full test suite → no regressions
3. If the bug was in a hot path, check performance
4. If the bug was a race condition, run with `-race` flag (Go) or equivalent

```
# Go example
go test ./... -count=1 -race

# Node example
npm test

# Python example
pytest -x -v
```

### Step 8: DOCUMENT

Save the finding to Engram for future reference:

```
mem_save(
  title: "Fixed [brief description]",
  type: "bugfix",
  content: "**What**: [what was fixed]\n**Why**: [root cause]\n**Where**: [files]\n**Learned**: [gotcha or pattern to watch for]"
)
```

## Error Analysis Patterns

### Stack Trace Reading

Read stack traces bottom-up (most languages):
1. Find the first frame in YOUR code (skip framework/stdlib frames)
2. That's your entry point for investigation
3. Read the error message — it often tells you exactly what's wrong
4. Check the line number — is it an off-by-one from the actual problem?

### Common Root Cause Categories

| Category | Symptoms | Investigation |
|----------|----------|---------------|
| **Nil/null reference** | Panic, NPE, "undefined is not a function" | Trace where the nil value originates |
| **Race condition** | Flaky tests, intermittent failures | Run with race detector, check shared state |
| **Off-by-one** | Wrong count, missing last element, index out of bounds | Check loop bounds, slice indices |
| **State mutation** | Works first time, fails on retry | Check for shared mutable state, missing reset |
| **Type mismatch** | Unexpected behavior, silent wrong results | Check type assertions, JSON unmarshaling |
| **Error swallowing** | Silent failure, missing data | Grep for `_ = err`, empty catch blocks |
| **Import/dependency** | Build fails, wrong version behavior | Check go.mod, package.json, lock files |
| **Environment** | Works locally, fails in CI | Check env vars, paths, permissions, OS differences |
| **Encoding** | Garbled text, wrong characters | Check UTF-8, BOM, line endings (CRLF vs LF) |
| **Timing** | Timeout, deadlock, slow response | Check async operations, connection pools, locks |

### Test Failure Triage

When a test fails:

1. Read the test name — what is it testing?
2. Read the assertion — what was expected vs actual?
3. Is this a new test or existing? (`git log --oneline <test-file>`)
4. Did the test ever pass? (`git log --all -p -S "TestName"`)
5. Is it flaky? (run 3 times — `go test -count=3 -run TestName`)

For flaky tests specifically:
- Check for time-dependent logic (use injectable clocks)
- Check for shared test state (use t.Parallel() carefully)
- Check for port conflicts (use random ports)
- Check for file system race conditions (use t.TempDir())

## Anti-Patterns

### Don't: Shotgun debug
```
# BAD — changing random things without understanding
// Maybe this is the problem?
-if (x > 0) {
+if (x >= 0) {
// Nope, try this
-if (x >= 0) {
+if (x > -1) {
// Still broken, try this...
```

### Do: Hypothesize then verify
```
# GOOD — state what you think is wrong, then test it
HYPOTHESIS: The comparison should be >= because x=0 is a valid input
EVIDENCE: The function spec says "non-negative integers"
PREDICTION: Adding x=0 to the test should fail with current code
→ Confirmed. Fix: change > to >=
```

### Don't: Fix the symptom
```
# BAD — wrapping in try/catch without understanding why it throws
try {
  processData(input);
} catch (e) {
  // ignore
}
```

### Do: Fix the root cause
```
# GOOD — understand why it throws and prevent it
// processData throws when input.items is undefined
// Root cause: API response omits items when empty instead of returning []
const items = input.items ?? [];
processData({ ...input, items });
```

### Don't: Debug in production
```
# BAD — adding console.log to production code
console.log("DEBUG: value is", value);  // committed to main
```

### Do: Use proper debugging tools
```
# GOOD — use debugger, targeted tests, or temporary diagnostic
// Write a failing test that captures the bug
func TestProcessData_EmptyItems(t *testing.T) {
    input := Input{Items: nil}
    result, err := ProcessData(input)
    // This test documents the bug and verifies the fix
    if err != nil {
        t.Fatalf("unexpected error for nil items: %v", err)
    }
}
```

## Performance Debugging

When investigating performance issues:

1. **Measure first** — get a baseline before changing anything
2. **Profile** — use language-specific profiling tools
   - Go: `pprof`, `trace`, `-benchmem`
   - Node: `--prof`, Chrome DevTools
   - Python: `cProfile`, `py-spy`
3. **Identify the bottleneck** — is it CPU, memory, I/O, network?
4. **Fix the biggest bottleneck first** — Amdahl's law
5. **Measure again** — confirm improvement with numbers

Anti-pattern: optimizing code that isn't the bottleneck.

## Integration with Other Skills

- After fixing a bug, if the change is non-trivial, trigger `@code-reviewer`
- If the bug involves security (auth bypass, injection, data leak), trigger `@security-reviewer`
- If the fix requires deployment, trigger `@release-manager`
- If the fix touches high-risk areas, load `op-guardrails` before applying
