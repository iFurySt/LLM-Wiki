# Installation

Install the package globally to make the `chrome-devtools` command available:

```sh
npm i chrome-devtools-mcp@latest -g
chrome-devtools status
```

## Repo Browser Profile

This repo reserves `.llmwiki/chrome-profile` for browser automation and manual verification.

- use Chrome stable with that profile
- do not point the CLI at a random fresh profile
- do not use the system default Chrome profile for repo verification
- if an existing stable instance already owns `.llmwiki/chrome-profile`, reuse it

Preferred repo entrypoints:

```sh
make browser-open
make browser-mcp
```

Equivalent manual launch shape when login or recovery is needed:

```sh
mkdir -p .llmwiki/chrome-profile
open -na "Google Chrome" --args \
  --user-data-dir="$(pwd)/.llmwiki/chrome-profile" \
  --remote-debugging-port=9222
```

After login or setup, keep reusing the same profile.

## Troubleshooting

- `chrome-devtools` not found: ensure the global npm bin directory is on `PATH`.
- Wrong browser instance: stop the stray session and reconnect to the Chrome stable instance that uses `.llmwiki/chrome-profile`.
- No login state: launch headed Chrome stable with `.llmwiki/chrome-profile`, finish login, then continue with the CLI.
