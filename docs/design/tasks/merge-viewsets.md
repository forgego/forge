Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# One viewset: BaseViewSet gains authentication/permissions/throttles; the broken "enhanced" stack is removed

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge (run go commands from there).
Only edit files in forge/api (not its subpackages except tests that reference removed types). Other agents edit forge/filter, forge/db, forge/identity, forge/registry, forge/cli, forge/config concurrently: ignore build errors there.

## Principles (mandatory)
- Keep every public name that forge/cli templates, forge/codegen/templates, examples/ or docs-site use: `api.ViewSet`, `api.BaseViewSet`, `api.NewBaseViewSet`, `api.ViewSetConfig`, `api.Router`/`NewRouter`, `ConfigurableViewSet`. Generated code must keep compiling.
- Behaviour of a `BaseViewSet` with no authentication/permission/throttle classes must not change.
- Tests FIRST for the new hooks. gofmt; functions < 40 lines; early returns; naming after behaviour.

## Facts (verified on master)
- Public viewset used by generated code, CLI scaffolds and examples: `BaseViewSet` (forge/api/viewset.go:51), built by `ViewSetConfig`/`ConfigurableViewSet` (viewset_config.go), used in forge/cli/templates/templates/api.go.tmpl, forge/codegen/templates/api.tmpl, forge/cli/commands/project/add_api.go:126, auth.go:156.
- `EnhancedBaseViewSet` (viewset_enhanced.go, 787 lines) duplicates BaseViewSet CRUD and adds authentication, permission, object-permission and throttle checks plus DRF-style exceptions. Its Create/Retrieve/Update/PartialUpdate/Destroy call `getManagerFromModel` (viewset.go:756), which always returns `reflect.Value{}`, so every one of them responds "Manager not found". Only List works.
- `EnhancedBaseViewSetIntegrated` (viewset_enhanced_integrated.go), `EnhancedRouter`/`ActionRegistry` (router_enhanced.go), `SetupCompleteAPI`/`CreateProductionViewSet`/`RegisterAPIWithDefaults`/`CompleteExample` (integration.go) and `CreateDefaultViewSet` (helpers.go:71) only exist to build the enhanced stack; nothing in examples, templates, CLI or docs-site uses them (confirm with grep across the workspace before deleting).

## Change
1. Add to `BaseViewSet` optional fields `Authentication []authentication.Authentication`, `Permissions []permissions.Permission`, `Throttles []throttling.Throttle`, and move the enhanced stack's `authenticateRequest`, `checkPermissions`, `checkObjectPermissions`, `checkThrottles` and `handleException` onto `BaseViewSet` (same logic, same exceptions). Call them at the start of List/Create/Retrieve/Update/PartialUpdate/Destroy, and the object-permission check after the instance is loaded in Retrieve/Update/PartialUpdate/Destroy. When the three slices are empty the checks are no-ops, so existing responses stay identical. Check whether `permissions.ViewSet` / action names need a `SetAction`/`GetAction` on BaseViewSet (the enhanced one had them); add the minimum needed.
2. Extend `ViewSetConfig` with the same three optional fields and pass them through in `NewConfigurableViewSet`.
3. Delete viewset_enhanced.go, viewset_enhanced_integrated.go, router_enhanced.go, integration.go, `getManagerFromModel`, and in helpers.go `CreateDefaultViewSet` (keep the other Get/Set default helpers if anything uses them; delete those with no users). Delete or port tests that only cover deleted code (viewset_enhanced_test.go, integration_test.go, router tests for EnhancedRouter). Port any test of permission/throttle behaviour to BaseViewSet.
4. Tests (new file viewset_access_checks_test.go): a BaseViewSet with a permission that denies → 403 on List and on Retrieve with the enhanced stack's response body; an authentication class that fails → 401; a throttle that denies → 429 with Retry-After if the enhanced code set it; no classes configured → unchanged 200 on List (use an in-memory SQLite queryset the way existing viewset tests do, or the existing fakes).

## Acceptance (from the module root)
    gofmt -l api                                      # empty
    go build ./... && (cd ../examples/ecommerce && go build ./...)
    go vet ./api/... ./cli/... ./codegen/...
    staticcheck ./api/...                             # no output
    go test -race ./api/... ./cli/... ./codegen/... -count=1

Report (max 30 lines): deleted files/identifiers, grep evidence of no users, moved checks, tests added/ported, each command result.
