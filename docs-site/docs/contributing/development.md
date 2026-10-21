---
sidebar_position: 32
description: Contributor entry point for Forge development.
---

# Development

Start with the canonical guide:

- [CONTRIBUTING on GitHub (master)](https://github.com/forgego/forge/blob/master/CONTRIBUTING.md)
- [Review standard](https://github.com/forgego/forge/blob/master/docs/REVIEWING.md)
- [Documentation policy](https://github.com/forgego/forge/blob/master/docs/DOCUMENTATION-POLICY.md)

## Validation summary

Use the Common Commands in CONTRIBUTING Quick Reference
(build, test, lint, frontend, docs) and keep docs links
and the docs-site build green after moves.

Use Go 1.26.9 or newer; CI tests 1.26.9 and 1.27.3. For frontend checks,
use Node 26.8.1 as CI does. Older Node releases must satisfy the installed
dependencies' engine requirements, not only Docusaurus's minimum.

The docs-site dependency overrides keep `tinypool` on 2.2.x or newer and
`postcss-selector-parser` on 7.1.6 or newer to avoid vulnerable versions
required by Docusaurus 3.10.2's dependency tree. Verify both the normal build
and Docusaurus's optional SSG worker-thread build when changing these
overrides. Do not use `npm audit fix --force`: its suggested Docusaurus
downgrade does not establish compatibility or resolve every finding.

The remaining `braces <=3.0.3` audit finding (GHSA-vfj7-8cjw-p6xm) has no
published fixed version as of 2026-10-20. It affects the docs development/build
dependency tree; the generated static site does not ship that Node runtime.
Keep the audit failure visible until a compatible upstream fix is available.
