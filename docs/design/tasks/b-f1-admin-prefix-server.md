TASK F1: serve the admin SPA correctly under a custom mount prefix (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/admin-mount-prefix/forge
(git worktree on branch fix/admin-mount-prefix, based on master.)

Files you may modify — ONLY:
  server/static.go            (new option + apply it)
  admin/site.go               (use the option)
  tests: server/static_test.go (existing or NEW), NEW admin/site_prefix_test.go
Do NOT edit go.mod / go.sum. Do NOT touch the React UI (task F2 does that).

=== PROBLEM ===
admin/site.go Handler() serves the built UI with
    server.StaticFS("", s.uiConfig.EmbedFS, server.WithPrefix(prefix), server.WithIndexFiles("index.html"),
                    server.WithFallback("index.html"), server.WithDisableCache(true))
(and StaticFiles for a local dir). The UI build hardcodes the "/admin/" base: index.html references
"/admin/assets/…", and the React app assumes "/admin" for its router basepath and API base. Mounting
the admin at any other prefix (uiConfig.Prefix, also passed to rest.Router.WithAdminPrefix) breaks
asset loading, routing, API calls and redirects.

The UI side (task F2) will read the runtime prefix from
    <meta name="forge-admin-prefix" content="/custom">
So the server must inject that tag and rewrite the "/admin/" asset URLs in index.html.

=== FIX ===
1. server/static.go: add an option
       // WithIndexTransform rewrites the bytes of index/fallback HTML files before they are served.
       func WithIndexTransform(fn func([]byte) []byte) StaticOption
   Read the existing StaticOption / config struct and the code paths that serve an index file or the
   fallback file. Apply fn ONLY when the served file is one of the configured index files or the fallback
   (i.e. HTML documents), never to other assets. Set Content-Length correctly for the transformed body
   (or omit it) and keep existing cache headers. If serving uses http.ServeContent/ServeFile, read the
   file bytes, transform, and write them with Content-Type text/html; charset=utf-8.

2. admin/site.go: compute
       prefix := normalized s.uiConfig.Prefix  ("" or "/" -> "", otherwise leading "/" and no trailing "/")
   and pass `server.WithIndexTransform(adminIndexTransform(prefix))` to BOTH StaticFiles and StaticFS.
   Add unexported:
       func adminIndexTransform(prefix string) func([]byte) []byte
   which:
     a. injects `<meta name="forge-admin-prefix" content="PREFIX">` immediately after the opening
        `<head>` tag (match `<head>` case-insensitively; if absent, prepend to the document).
        PREFIX must be HTML-attribute escaped (html.EscapeString).
     b. replaces every occurrence of `"/admin/` with `"PREFIX/` and `'/admin/` with `'PREFIX/`
        ONLY when prefix != "/admin" (i.e. rewrite the build's default base to the runtime one).
        For prefix "" the result is `"/` .
   Keep it a pure function (unit-testable).

3. Before writing code, read rest.Router.RegisterRoutes to confirm the API is mounted relative to the
   same prefix (the UI will call `${prefix}/api/...`). Report what you find; do not change routing unless
   the API is demonstrably NOT reachable at `${prefix}/api` when the site is mounted at prefix.

=== TESTS ===
admin/site_prefix_test.go (table test of adminIndexTransform):
  - prefix "/admin": meta content "/admin" injected; asset URLs unchanged.
  - prefix "/backoffice": `src="/admin/assets/index-x.js"` -> `src="/backoffice/assets/index-x.js"`,
    `href="/admin/assets/a.css"` -> `href="/backoffice/assets/a.css"`; meta content "/backoffice".
  - prefix "": asset -> `src="/assets/index-x.js"`; meta content "".
  - prefix with a quote character is escaped in the meta tag.
  - document without <head> still gets the meta tag.
server static test: serving a StaticFS (fstest.MapFS with index.html and assets/app.js) with
WithFallback("index.html") and WithIndexTransform(upper-casing fn):
  - GET "/"            -> transformed body
  - GET "/deep/route"  -> fallback, transformed body
  - GET "/assets/app.js" -> original bytes (NOT transformed)

=== VERIFY ===
  gofmt -l server admin     (your files must not appear)
  go vet ./server/... ./admin/...
  go test -race ./server/ ./admin/...
  go test ./...   (TestPrefetchRelated_Integration in ./orm is known-flaky)
  cd ../examples/ecommerce && go build ./... && cd -

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- Additive API only (WithIndexTransform). Minimal diff.
- Final report: what RegisterRoutes mounts, files changed, tests, tail of the -race run.
