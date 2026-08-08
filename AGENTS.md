# Repository guide

This is a Hugo site for Go Meetup Prague.

## Where things live

- Site configuration and current meetup details: `config.toml`
- Homepage and page templates: `themes/netlify-basic/layouts/`
- Site styles: `themes/netlify-basic/static/css/demo-styling.css`
- Browser tests: `tests/site.spec.js`
- Video archive data: `data/videos.json`
- Video update tooling: `scripts/`

Keep changes small and reuse the existing Hugo, HTML, and CSS patterns. Do not add a dependency when the platform or current code already covers the need.

## Local preview

Always render the site through Hugo. Opening a template with `file://` does not process Hugo templates and breaks absolute asset paths.

```sh
hugo server --bind 127.0.0.1 --port 4173 --disableFastRender
```

Open <http://127.0.0.1:4173/> and leave the server running while the user reviews the result.

Before starting another server, check whether port 4173 is already in use:

```sh
lsof -nP -iTCP:4173 -sTCP:LISTEN
```

Reuse a server only after confirming that it belongs to this checkout and serves current files. Do not silently kill an unknown process. A stale server can make Playwright test old HTML.

## Current meetup banner

The homepage banner reads `[params.nextMeetup]` from `config.toml`:

- `url`: direct Meetup event URL
- `date`, `time`, and `place`: human-readable English display values
- `endsAt`: ISO 8601 timestamp with timezone; browser JavaScript hides the banner after this time

Update all five values when publishing a new meetup.

## Validation

Run the smallest relevant checks before pushing:

```sh
hugo
git diff --check
npm test
```

Stop a preview server started by the agent before `npm test`, then restart it for review. Playwright otherwise reuses anything already listening on port 4173. If files under `scripts/` change, also run:

```sh
cd scripts && go test ./...
```

## Git and pull requests

Every change, including documentation, content, and configuration, goes through a pull request for another review.

1. Start from an up-to-date `main`.
2. Create a focused `codex/<description>` branch.
3. Stage only files that belong to the requested change.
4. Commit, push, and open a draft PR against `main`.
5. Include the change, reason, user impact, and validation in the PR body.
6. Do not merge the PR unless the user explicitly asks; leave the final review and merge to the user or the next reviewer.

This machine uses the personal GitHub account from the keyring. Herdr's shared configuration may inject `GH_TOKEN` for multi-account setups and override the keyring. Prefer:

```sh
env -u GH_TOKEN gh auth status
env -u GH_TOKEN gh <command>
```
