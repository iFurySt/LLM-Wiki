# Harness Engineering Notes

Source:

- OpenAI, "Harness engineering: leveraging Codex in an agent-first world", published February 11, 2026: https://openai.com/index/harness-engineering/

These are distilled repo-practice notes, not a copy of the article.

## Extracted Principles

- Keep `AGENTS.md` short and use it as a map into the repo.
- Treat repository-local docs as the system of record.
- Prefer progressive disclosure over one giant instruction file.
- Make plans first-class and versioned in-repo.
- Optimize the repo for agent legibility, not only for human convenience.
- Push more durable knowledge into versioned markdown instead of chat or external docs.
- Favor boring, inspectable, well-understood technology during early agent-heavy development.

## How This Applies To LLM-Wiki

- `AGENTS.md` should stay compact.
- `docs/` should be structured and discoverable.
- stable knowledge, plans, todos, and decisions should not live in one file
- future agents should be able to infer repo intent from repository-local artifacts alone
