# PR marathon workflow

A **PR marathon** is the repository's reusable workflow for giving a coding agent a sequence of related pull-request goals that it can implement autonomously with the `goal` skill, without requiring a new human prompt between each PR.

The marathon automates implementation and PR preparation. It does **not** transfer integration authority to the coding agent.

`AGENTS.md` remains authoritative for repository-wide agent rules. In particular, coding agents are not authorized to merge PRs unless the user explicitly grants merge authority for a specific PR or action.

## Why this workflow exists

Some changes are easier to review and reason about as several focused PRs even when they form one larger implementation effort. A marathon lets an agent keep working through that sequence without turning the work into one giant branch or requiring the user to manually kick off each PR.

The intended separation is:

```text
human / planning agent
  -> define ordered PR goals
  -> package implementation prompts
  -> provide one kickoff prompt

coding agent using goal skill
  -> implement PR 1
  -> open/update PR 1
  -> implement PR 2
  -> open/update PR 2
  -> ...
  -> leave the stack review-ready

human
  -> review
  -> authorize or perform merges
```

## Marathon artifacts

A marathon handoff consists of two distinct artifacts.

### 1. Prompt package

Write one self-contained Markdown implementation prompt per PR.

Use numeric filenames so order is unambiguous, for example:

```text
01-domain-foundation.md
02-runtime.md
03-provider.md
04-integration.md
```

Package those Markdown files in a ZIP archive.

The ZIP should contain only the queued implementation prompts and any deliberately included supporting handoff material. It must not contain secrets.

### 2. Kickoff prompt

The initial chat prompt is separate from the ZIP. Do not include it inside the prompt package.

The kickoff prompt should tell the coding agent to:

- use the `goal` skill for the entire marathon;
- inspect the attached ZIP rather than assuming a local folder path;
- extract the ZIP outside the repository or into an explicitly untracked/gitignored temporary location;
- read all queued prompts before implementation so it understands the dependency chain;
- execute the prompts in numeric order;
- use separate worktrees and purpose-based branches;
- continue through the queue without routine human intervention;
- obey the merge-authorization boundary in `AGENTS.md` and this document;
- stop only for a genuine blocker that cannot safely be resolved from repository state, documentation, tests, review feedback, or normal engineering judgment.

## Hard merge boundary

A PR marathon does **not** authorize merges.

The following do not constitute merge permission:

- "complete the marathon";
- "finish all PRs";
- "work through the queue";
- "use the goal skill";
- broad autonomy to fix tests or review findings;
- repository permissions that technically allow the agent to merge.

Unless the user explicitly authorizes a particular merge, the agent must leave PRs open for human review.

The coding agent must not:

- merge a PR;
- enable auto-merge;
- push implementation commits directly to `main`;
- close a PR as a substitute for integration;
- self-approve on the user's behalf;
- bypass branch protection or review requirements;
- treat a passing CI run as equivalent to human approval.

A coding agent may:

- create branches/worktrees;
- commit and push its implementation;
- open PRs;
- update its PR branches;
- run and repair CI/test failures;
- respond to review comments;
- resolve a review conversation after its underlying finding has actually been addressed;
- rebase or otherwise maintain its own feature branches when safe and consistent with repository policy.

## Dependent PRs: use a stack

Because the agent cannot merge predecessor PRs itself, dependent marathon PRs should normally be created as a **stack**.

For a four-PR marathon:

```text
main
  |
  +-- PR 1 branch
        |
        +-- PR 2 branch
              |
              +-- PR 3 branch
                    |
                    +-- PR 4 branch
```

Open the PRs with matching bases:

```text
PR 1: base main
PR 2: base PR 1 branch
PR 3: base PR 2 branch
PR 4: base PR 3 branch
```

This allows the coding agent to continue through the entire marathon without integrating unreviewed work into `main`.

Each PR should contain only the incremental diff for its own goal relative to its immediate predecessor branch.

### After human merges

Review and merge from the bottom of the dependency chain upward: PR 1 first, then PR 2, and so on.

After a predecessor is merged, the next PR should be retargeted to `main` and its branch updated/rebased if necessary so the PR shows only its intended incremental changes.

**The retarget step is not optional, and skipping it silently loses work.** A pull request merges into its base branch. GitHub retargets a child to `main` only once its parent has actually merged, and that retarget does not reliably land if the stack is merged in quick succession. Every pull request whose base is still the parent branch then merges into a branch `main` never sees again — while reporting as merged, with no failing check, no conflict, and no broken build.

This has happened twice in this repository, orphaning eight pull requests across two stacks. In both cases only the pull request whose base was already `main` actually landed.

Two safe procedures, either of which avoids it:

- **Merge only the top branch of the stack.** For a linear stack the top branch already contains every commit below it, so one merge integrates the whole marathon. The lower pull requests then close as merged on their own.
- **Or retarget each pull request to `main` before merging it, one at a time**, confirming the retarget took effect before moving to the next.

After any stack merge, confirm the work actually arrived:

```bash
git merge-base --is-ancestor "$(gh pr view <number> --json headRefOid --jq .headRefOid)" origin/main
```

The `Merge topology` workflow enforces both halves of this automatically: it fails a pull request whose base branch has already been merged, and audits `main` for merged work that is unreachable from it. Note that a branch force-pushed *after* its pull request merged has to be recorded in `.github/reconciled-pull-requests.txt`, because GitHub then reports a rebased tip that `main` has never seen even though the original commits landed.

An agent may perform that branch maintenance if asked or if it remains active during the review phase, but it still may not perform the merge itself without explicit authorization.

If stacking is impractical for a particular repository/tool limitation, the marathon must pause at the dependency boundary rather than silently merge a predecessor.

## Worktree rules

Every implementation PR should normally use its own worktree.

Do not implement marathon work directly in the main checkout.

For dependent stacked PRs, create the next branch/worktree from the predecessor branch rather than from stale `main`.

Do not let multiple agents mutate the same working tree.

When a PR branch is no longer needed locally, clean up its worktree only when doing so cannot destroy uncommitted work.

## Branch, commit, and PR naming

Names describe the purpose of the change, not the agent identity.

Use repository-standard prefixes such as:

- `feat/...` for feature branches;
- `fix/...` for bug fixes;
- `docs/...` for documentation-only work.

Commit messages and PR titles should use corresponding conventional prefixes such as:

- `feat: ...`;
- `fix: ...`;
- `docs: ...`;
- `test: ...` where genuinely appropriate.

Do not use prefixes such as `agent/`, `codex/`, an agent name, or author identity.

## What each PR prompt should contain

Each Markdown prompt should be sufficiently self-contained for implementation while acknowledging its place in the marathon.

At minimum include:

- PR goal and reason;
- expected branch name;
- dependency on prior marathon PRs;
- repository/docs that must be read;
- required behavior;
- architectural constraints;
- important implementation boundaries;
- explicit non-goals;
- tests and validation required;
- acceptance criteria;
- any live-test credentials or external resources, described without embedding secrets;
- the expected handoff/end state for the next PR.

Prompts should specify behavior and boundaries without forcing implementation details that conflict with the current codebase. The coding agent must inspect current repository state before coding.

## Review behavior during a marathon

The agent should keep each opened PR healthy while continuing the queue.

If automated CI or review feedback arrives while later PRs are being implemented, the agent may return to the affected PR/worktree and address actionable findings, provided dependency relationships are maintained safely.

When fixing a lower PR in the stack, propagate/update descendant branches as necessary so later PRs include the correction without duplicating or reverting it.

Do not resolve a review conversation merely to make the PR appear clean. Resolve it only after the finding has actually been addressed or the reviewer/user has explicitly accepted another disposition.

A coding agent must not submit an approval on behalf of the user.

## Secrets and attached archives

Marathon ZIP files and extracted prompt files are temporary orchestration artifacts, not repository content.

They must be extracted outside the repository or into an explicitly untracked/gitignored temporary location and must never enter Git history.

Secrets must never be included in the ZIP or Markdown prompts.

If a local untracked file is supplied as a credential source for smoke tests:

- leave it untracked;
- do not copy it into feature worktrees or tracked `.env` files;
- do not print/log its contents;
- source it directly into the test process environment when needed;
- keep committed tests independent of the live credential.

## Marathon completion

Without explicit merge authorization, **marathon completion means implementation-complete and review-ready, not merged**.

A marathon is complete when:

- every queued PR has been implemented;
- every PR has been opened with the correct base/head relationship;
- required formatting, tests, generators, and CI are passing or any genuine external blocker is clearly documented;
- known actionable review findings have been addressed;
- review conversations that were actually addressed have been resolved where appropriate;
- no secrets, prompt archives, or unrelated files were committed;
- the complete PR stack is ready for human review and integration.

The final agent report should list:

- each PR number/title/link;
- branch and base relationship;
- validation/CI status;
- important implementation decisions;
- any sanitized live-smoke-test results;
- review findings addressed;
- remaining caveats or deliberately deferred work;
- a clear statement that the PRs remain unmerged unless the user explicitly authorized specific merges.

## Reusing the workflow

When the user asks to "package this as a PR marathon" or otherwise invokes the established marathon workflow, planning agents should treat this document as the source of truth rather than reconstructing the process from conversational memory.

If the user explicitly changes a marathon rule for a particular run, that instruction governs that run. Durable changes to the workflow should be added to this document so future marathons do not depend on remembered chat context.
