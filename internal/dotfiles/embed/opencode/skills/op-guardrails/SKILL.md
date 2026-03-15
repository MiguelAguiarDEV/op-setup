---
name: op-guardrails
description: >
  Operational guardrails for high-risk actions. Trigger: sensitive data access,
  external side effects (network, APIs, payments), tool/package installation,
  destructive or irreversible changes, rollback decisions, failure triage,
  production environment access, secret/credential handling, or any action
  where the blast radius is non-trivial.
---

# Operational Guardrails

Governance layer for side-effecting, destructive, or high-risk operations.
This skill activates when the agent is about to do something that could cause
real-world damage, data loss, or security exposure on the user's machine.

The goal is not to block work — it is to ensure every risky action is
deliberate, reversible when possible, and auditable after the fact.

## When to Activate

Load this skill when any of these conditions are true:

- **File system mutations** outside the current project directory
- **Package/tool installation** (npm i -g, brew install, go install, pip install, cargo install)
- **Git destructive ops** (force push, hard reset, rebase on shared branches, branch deletion)
- **Secret/credential access** (reading .env, credentials.json, tokens, API keys)
- **External API calls** (HTTP requests, webhook triggers, cloud CLI commands)
- **Database operations** (migrations, seed, truncate, drop)
- **System configuration changes** (shell rc files, PATH modifications, systemd units)
- **Docker/container ops** (image builds, volume mounts, network changes)
- **CI/CD pipeline triggers** (deploy, release, publish)
- **Rollback or recovery** after a failed operation
- **Bulk operations** affecting more than 10 files or records

Do NOT activate for:
- Read-only exploration (grep, find, cat, git log, git status)
- Edits within the current project that are easily reversible (git checkout)
- Running existing test suites
- Local dev server start/stop

## Risk Classification

Before executing any guarded action, classify it:

| Level | Criteria | Gate |
|-------|----------|------|
| **LOW** | Reversible, local scope, no external systems | Log intent, execute |
| **MEDIUM** | Partially reversible, touches config or shared state | Propose + confirm |
| **HIGH** | Irreversible, external systems, production, secrets | Full proposal + explicit confirm |
| **CRITICAL** | Data loss, security exposure, production deploy | Full proposal + explicit confirm + verification plan |

## Execution Proposal Format

Before any MEDIUM+ action, present this to the user:

```
# Execution Proposal

## Risk: [LOW | MEDIUM | HIGH | CRITICAL]

## Intent
- What: [one sentence — what will happen]
- Why: [one sentence — why this is needed]

## Actions (batched)
1. [exact command or operation]
2. [exact command or operation]

## Files/Systems Touched
- [path or system name]

## Reversibility
- [How to undo this, or "IRREVERSIBLE" if not possible]

## Risks
- [What could go wrong]

Confirm? [y/n]
```

For LOW risk: log the action inline and proceed without blocking.

## Tool Installation Protocol

When a tool is missing and needed:

1. State what tool is needed and why
2. Propose the exact install command
3. State the scope (global vs local, user vs system)
4. Wait for confirmation
5. Install
6. Verify installation succeeded
7. Resume the original workflow

Anti-pattern: silently installing tools or retrying with different package managers.

```
# Example
Tool `fd` not found. Required for fast file search.
Install: `cargo install fd-find` (user-level, ~/.cargo/bin)
Alternative: fall back to `find` (slower, no .gitignore awareness)
Confirm?
```

## Secret Handling Rules

1. Never read secret files (.env, credentials.json, *_key, *_token, *.pem) without explicit user request
2. Never echo, log, or display secret values in output
3. Never commit files that likely contain secrets — warn the user
4. Never pass secrets as command-line arguments (visible in process list)
5. When secrets are needed, prefer environment variable references over literal values
6. If a secret is accidentally exposed in output, immediately flag it

## Git Safety Protocol

These rules apply to all git operations:

| Operation | Risk | Rule |
|-----------|------|------|
| `git push --force` | CRITICAL | Warn. Never force-push to main/master without explicit request |
| `git reset --hard` | HIGH | Confirm. State what commits/changes will be lost |
| `git rebase` on shared branch | HIGH | Confirm. Explain divergence risk |
| `git branch -D` | MEDIUM | Confirm. State if branch has unmerged commits |
| `git clean -fd` | HIGH | Confirm. List files that will be deleted |
| `git checkout -- .` | MEDIUM | Confirm. State uncommitted changes that will be lost |
| `git commit --amend` after push | CRITICAL | Warn. Requires force push |

## Failure Triage Protocol

When an operation fails:

1. **Stop** — do not retry automatically
2. **Diagnose** — read the error, identify root cause
3. **Classify** — is this recoverable? What is the current state?
4. **Report** — present findings to user with options:
   - Option A: fix and retry
   - Option B: rollback to previous state
   - Option C: abort and document the failure
5. **Wait** for user decision before acting

Anti-pattern: retrying failed operations in a loop hoping they succeed.

## Rollback Decision Framework

When deciding whether to rollback:

```
Is the system in a consistent state?
├── YES → Document partial progress, continue or pause
└── NO → Is rollback possible?
    ├── YES → Propose rollback with exact steps
    └── NO → STOP. Present current state and options to user.
```

Rollback steps must be presented as an Execution Proposal (MEDIUM+ risk).

## Bulk Operation Safety

For operations affecting >10 files or records:

1. Show a preview of affected items (first 5 + count)
2. State the total count
3. Confirm before executing
4. Execute in batches if possible (allows partial abort)
5. Report results with success/failure counts

## Audit Trail

After any MEDIUM+ operation completes:

1. Save to Engram with type `decision` or `config`:
   ```
   mem_save(
     title: "Installed X / Modified Y / Deployed Z",
     type: "config",
     content: "**What**: ...\n**Why**: ...\n**Where**: ...\n**Learned**: ..."
   )
   ```
2. Include the risk level and whether rollback was needed

## Anti-Patterns

### Don't: Execute and apologize
```
# BAD — running destructive command then checking if it was ok
rm -rf node_modules
# "Oops, did you need those?"
```

### Do: Propose and confirm
```
# GOOD — present what will happen, wait for confirmation
# Execution Proposal
# Risk: LOW
# Intent: Remove node_modules to force clean install
# Reversibility: `npm install` restores from package-lock.json
# Confirm?
```

### Don't: Retry blindly
```
# BAD — retrying a failed deploy without understanding why
npm run deploy  # failed
npm run deploy  # retry
npm run deploy  # retry again
```

### Do: Diagnose first
```
# GOOD — understand the failure before acting
# Deploy failed: exit code 1
# Error: "Authentication token expired"
# Root cause: CI token needs refresh
# Options:
#   A) Refresh token and retry
#   B) Abort and investigate token rotation
# Waiting for decision.
```
