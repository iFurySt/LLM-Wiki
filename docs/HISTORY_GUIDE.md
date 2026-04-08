# History Guide

This repo uses `docs/worklog/`, `docs/decisions/`, and `docs/test-results/` as durable change history.

Read this document before commit.

## What To Update

- Update `docs/worklog/` for milestones, shipped capabilities, or meaningful repo-level progress.
- Update `docs/decisions/` for significant product, architecture, protocol, auth, install, or workflow decisions.
- Update `docs/test-results/` when a test run produces durable evidence worth keeping for future comparison.

## Worklog Rules

- Append, do not rewrite history except to fix factual mistakes.
- Keep entries factual and outcome-oriented.
- Group entries under the correct month file such as `docs/worklog/2026-04.md`.
- Describe what changed and why it matters, not every intermediate step.

## Decision Rules

- Record the decision, rationale, and replacement path if an older direction is being superseded.
- Prefer a dedicated dated decision document when the topic needs standalone context.
- Keep monthly index files current when new decision records are added.

## Test Result Rules

- Store concise summaries, not raw noisy logs.
- Capture the date, scope, command or surface tested, and the result.
- Only keep runs that are reusable as future evidence.

## Commit Checklist

1. Check whether this change creates a milestone worth logging in `docs/worklog/`.
2. Check whether this change makes or replaces a durable decision that belongs in `docs/decisions/`.
3. Check whether validation results should be preserved in `docs/test-results/`.
4. If none apply, commit without forcing a history update.
