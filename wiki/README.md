# Wiki sources

This directory is the source of truth for the project's GitHub wiki at
<https://github.com/dipjyotimetia/openapi-go-mcp/wiki>.

GitHub wikis are stored in a **separate git repo** (`<repo>.wiki.git`). You can
edit pages here in the main repo (where they go through normal PR review), then
push them to the wiki repo with the included `sync.sh` script.

## One-time bootstrap

GitHub doesn't create the wiki repo until the first page is published through
the web UI. Bootstrap it once:

1. Go to <https://github.com/dipjyotimetia/openapi-go-mcp/wiki>.
2. Click **Create the first page**. Save any placeholder content.
3. After that the wiki repo exists at
   `https://github.com/dipjyotimetia/openapi-go-mcp.wiki.git` and `sync.sh`
   can push to it.

## Syncing

```bash
./wiki/sync.sh
```

The script clones the wiki repo into `/tmp`, copies every `*.md` from
`wiki/` over it, commits, and pushes. Authentication uses whatever method
your existing `git push` uses (HTTPS token, SSH key, `gh auth`).

To preview the commit before pushing, run with `DRY_RUN=1`:

```bash
DRY_RUN=1 ./wiki/sync.sh
```

## Conventions

- **Page slugs use hyphens**, not spaces: `Getting-Started.md` ↔ wiki page
  "Getting Started". GitHub maps these automatically.
- **`_Sidebar.md` and `_Footer.md`** are reserved by GitHub for navigation.
- **Internal links** between wiki pages use `[Title](Page-Name)` (no `.md`).
- **Links into the source repo** use full GitHub URLs so they work both in
  this directory and on the rendered wiki.

## Why mirror instead of edit-in-place on GitHub?

So wiki edits go through code review like everything else, and so the docs
move with branches. The wiki on GitHub is always the latest snapshot from
`main`.
