---
name: op-skill-creator
description: Create new skills, modify existing skills, and validate skill quality. Trigger: When the user wants to create a skill from scratch, edit or improve an existing skill, validate a skill's structure, or understand skill design patterns.
---

# Skill Creator

A skill for creating, iterating, and validating other skills within the Baymax/OpenCode ecosystem.

Skills follow the [Agent Skills](https://agentskills.io) open standard. Each skill is a directory
with a `SKILL.md` file containing YAML frontmatter and markdown instructions.

## Core Loop

```
1. Capture intent
2. Interview + research
3. Write SKILL.md draft
4. Create test cases (2-3 realistic prompts)
5. Run skill on test cases, evaluate results
6. Iterate based on feedback
7. Register in AGENTS.md
```

Your job is to figure out where the user is in this process and help them progress.
If they say "I want a skill for X", start from step 1. If they already have a draft, jump to step 4.
If they say "just vibe with me", skip the formal eval and iterate conversationally.

---

## Step 1: Capture Intent

Extract from the conversation or ask directly:

1. What should this skill enable the agent to do?
2. When should this skill trigger? (what user phrases, file types, or contexts)
3. What is the expected output format or behavior?
4. Does the skill need supporting files (scripts, templates, references)?

If the current conversation already contains a workflow the user wants to capture,
extract answers from the conversation history first — tools used, sequence of steps,
corrections made, input/output formats observed. Confirm before proceeding.

## Step 2: Interview and Research

Proactively ask about:
- Edge cases and failure modes
- Input/output formats and examples
- Success criteria (how do we know the skill works?)
- Dependencies (tools, MCPs, other skills)
- Scope boundaries (what should the skill NOT do?)

If useful, search Engram for prior decisions or patterns related to the skill's domain:
```
mem_search("topic relevant to the skill")
```

## Step 3: Write the SKILL.md

### Anatomy of a Skill

```
skill-name/
├── SKILL.md           # Main instructions (required, <500 lines ideal)
├── references/        # Docs loaded into context as needed (optional)
├── assets/            # Templates, schemas (optional)
└── scripts/           # Executable helpers (optional)
```

### Required Frontmatter

```yaml
---
name: lowercase-hyphenated-name
description: >
  What this skill does and when to use it. Include specific trigger contexts.
  Be slightly "pushy" — Claude tends to under-trigger skills.
---
```

Only `name` and `description` are required. Additional fields for Claude Code compatibility:

| Field | Purpose |
|-------|---------|
| `disable-model-invocation` | `true` = only manual `/name` invocation, no auto-trigger |
| `user-invocable` | `false` = hidden from `/` menu, only agent uses it as context |
| `allowed-tools` | Restrict tools when skill is active (e.g., `Read, Grep, Glob`) |
| `context` | `fork` = run in isolated subagent with fresh context |
| `agent` | Subagent type when `context: fork` (e.g., `Explore`, `Plan`) |

### Progressive Disclosure (3 levels)

1. **Metadata** (name + description) — always in context (~100 words max)
2. **SKILL.md body** — loaded when skill triggers (<500 lines ideal)
3. **Bundled resources** — loaded on demand (unlimited size)

If SKILL.md approaches 500 lines, split into references/ with clear pointers:
```markdown
For complete API patterns, see [references/api-patterns.md](references/api-patterns.md)
```

### Writing Principles

These come directly from Anthropic's skill engineering research:

**Explain the why, not just the what.**
LLMs are smart. They have good theory of mind. When given reasoning, they go beyond
rote instructions. If you find yourself writing ALWAYS or NEVER in all caps, that is
a yellow flag — reframe and explain the reasoning so the model understands why.

**Generalize from examples.**
Skills will be used across many different prompts. If the skill works only for the
examples you tested with, it is useless. Use different metaphors, recommend different
patterns. Make instructions general enough to transfer.

**Keep the prompt lean.**
Remove things that are not pulling their weight. Read transcripts of test runs — if
the skill makes the model waste time on unproductive steps, cut those instructions.

**Use imperative form.**
"Run the test suite" not "You should run the test suite".

**Include examples with input/output pairs.**
```markdown
## Example
Input: Added user authentication with JWT tokens
Output: feat(auth): implement JWT-based authentication
```

**Include anti-patterns.**
Explicit "don't do this" sections with code are more actionable than positive-only instructions.

### Naming Convention

Skills in this ecosystem use prefixed names by domain:

| Prefix | Domain | Examples |
|--------|--------|----------|
| `fe-` | Frontend | `fe-react`, `fe-design`, `fe-baseline-ui` |
| `qa-` | Quality/Testing | `qa-debugging`, `qa-error-handling` |
| `op-` | Operations | `op-guardrails`, `op-config-sync` |
| `sec-` | Secretary/Services | `sec-gmail`, `sec-notion`, `sec-calendar` |
| `bsd-` | BySidecar-specific | `bsd-release-deploy` |

## Step 4: Create Test Cases

After writing the draft, create 2-3 realistic test prompts — what a real user would say.
Share them with the user for confirmation before running.

Good test cases:
- Cover the happy path and at least one edge case
- Use realistic language (not abstract requests)
- Include context a real user would provide

```
Test 1: "I need a skill that reviews my PR descriptions for completeness"
Test 2: "Create a skill for generating changelog entries from git commits"
Test 3: "Make a skill that enforces our API naming conventions"
```

## Step 5: Run and Evaluate

For each test case:

1. Read the skill's SKILL.md
2. Follow its instructions to accomplish the test prompt
3. Evaluate: did the output match expectations?
4. Note what worked and what did not

If subagents are available, run test cases in parallel for speed.

### Evaluation Criteria

| Criterion | Question |
|-----------|----------|
| **Correctness** | Does the output match the expected behavior? |
| **Completeness** | Are all requirements addressed? |
| **Consistency** | Does it produce similar quality across different inputs? |
| **Efficiency** | Does it avoid unnecessary steps or token waste? |
| **Generalization** | Does it work beyond the specific test cases? |

## Step 6: Iterate

Based on evaluation:

1. Identify patterns in failures (not just individual fixes)
2. Rewrite the skill addressing root causes
3. Re-run test cases
4. Repeat until satisfied

Focus on understanding WHY something failed, not just patching symptoms.

## Step 7: Register

After the skill is validated:

1. Ensure the skill directory is in `~/.config/opencode/skills/<skill-name>/`
2. Update `AGENTS.md` skills table:
   ```markdown
   | `skill-context` | `skills/<skill-name>/SKILL.md` | Description |
   ```
3. If `skills/ACTIVE_SKILLS.txt` exists, add the skill name to it
4. Save to Engram:
   ```
   mem_save(
     title: "Created skill: <name>",
     type: "pattern",
     topic_key: "skills/<skill-name>",
     content: "**What**: Created <name> skill\n**Why**: <reason>\n**Where**: skills/<name>/SKILL.md\n**Learned**: <any gotchas>"
   )
   ```

---

## Modifying Existing Skills

When improving an existing skill:

1. Read the current SKILL.md completely
2. Identify the specific problem (triggering? output quality? scope?)
3. Make minimal changes — do not rewrite unless necessary
4. Test with the same cases that exposed the problem
5. Verify existing functionality is not broken

### Skill Patch Format

For targeted fixes, use this format to communicate changes:

```markdown
# Skill Patch: <skill-name>

## Symptom
- What is failing or suboptimal

## Root Cause
- Why it is happening

## Patch (minimal diff)
- Before: <relevant section>
- After: <changed section>

## Verification
- How to confirm the fix works
```

---

## Reference: Full SKILL.md Template

```yaml
---
name: template-skill
description: >
  What this skill does. When to use it. Include trigger contexts.
  Be specific about user phrases that should activate this skill.
---

# Skill Name

Brief overview of what this skill does and why it exists.

## When to Use

- Context 1 that should trigger this skill
- Context 2
- Context 3

## Core Patterns

### Pattern 1: Name
Explanation of why this pattern matters.

```language
// Code example
```

### Pattern 2: Name
...

## Anti-Patterns

### Don't: Description
```language
// Bad example
```

### Do: Description
```language
// Good example
```

## Examples

**Example 1:**
Input: <realistic input>
Output: <expected output>

**Example 2:**
Input: <realistic input>
Output: <expected output>
```
