# Forge Documentation Site

User documentation website for Forge (`https://forgego.github.io/forge/`), built with Docusaurus.

## Quick Start

```bash
npm ci
npm run generate-artifacts   # generates static/llms.txt, static/llms-full.txt, static/search-index.json
npm start                    # starts dev server on http://localhost:3000
```

## Build

```bash
npm run build
```

`npm run build` automatically runs `scripts/generate-site-artifacts.mjs` before `docusaurus build --no-minify`, producing `llms.txt`, `llms-full.txt`, and `search-index.json` inside `static/` (which Docusaurus copies into `build/`).
