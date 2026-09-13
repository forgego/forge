Your workspace is /home/hamid/Other/projects/foreit-wt/remove-unreachable. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Remove three packages that nothing imports and that are unreachable or harmful

Module root: /home/hamid/Other/projects/foreit-wt/remove-unreachable/forge. Delete only these directories (use `rm -r`), change nothing else:

1. forge/cli/internal: an `internal` package that no package under forge/cli imports, so no code (and no user) can ever reach it.
2. forge/filter/widgets: nothing imports it. `AutosuggestWidget.Render` writes the unescaped name, value and suggestions into HTML and into a `<script>` (autosuggest.go:28, :39), which is an XSS risk.
3. forge/db/migrate/dependencies: nothing imports it. It duplicates the dependency detector the generator actually uses (forge/db/migrate/generate/dependencies.go `DependencyDetector`).

Before deleting, confirm each has no importer:
    grep -rn --include=*.go --include=*.tmpl 'forge/cli/internal"\|forge/filter/widgets"\|forge/db/migrate/dependencies"' /home/hamid/Other/projects/foreit-wt/remove-unreachable
Expect no output. If any importer exists, STOP and report it without deleting anything.

## Acceptance (from the module root)
    go build ./...
    go vet ./cli/... ./filter/... ./db/...
    go test ./cli/... ./filter/... ./db/... -count=1
    (cd /home/hamid/Other/projects/foreit-wt/remove-unreachable/examples/ecommerce && go build ./...)

Report (max 10 lines): the importer grep result, directories deleted, each command result.
