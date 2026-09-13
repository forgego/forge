Your workspace is /home/hamid/Other/projects/foreit-wt/rename-tests. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Rename test files and tests that are named after a ticket ("w0")

Module root: /home/hamid/Other/projects/foreit-wt/rename-tests/forge. Pure rename/move: do not change any test logic, assertion or non-test file.
Create the new file with the same content, then delete the old file (use `rm`, not git).

## Renames
1. forge/api/errors/w0_not_implemented_test.go → forge/api/errors/database_store_not_implemented_test.go
2. forge/cli/commands/migrations/w0_not_implemented_test.go → forge/cli/commands/migrations/squash_not_implemented_test.go
3. forge/db/migrate/generate/w0_not_implemented_test.go → forge/db/migrate/generate/squash_not_implemented_test.go
4. forge/log/w0_not_implemented_test.go → forge/log/remote_output_not_implemented_test.go
5. forge/orm/w0_orm_test.go is split by behaviour (keep the package clause and only the imports each file needs):
   - forge/orm/queryset_not_implemented_test.go: `TestW0_SetOperations` → `TestQuerySet_SetOperationsNotImplemented`, `TestW0_Aggregate` → `TestQuerySet_AggregateNotImplemented`.
   - forge/orm/queryset_default_order_test.go: `TestW0_FirstLastDefaultPKOrdering` → `TestQuerySet_FirstLastDefaultToPrimaryKeyOrder`.
   - forge/orm/update_builder_set_nil_test.go: `TestW0_UpdateBuilderSetNil` → `TestUpdateBuilder_SetNil`.
   - The shared `TestPointerCustomer` type and its methods (top of the file) and any other helpers go to whichever new file uses them; if several use them, put them in queryset_default_order_test.go. Do not duplicate declarations.
   - Delete forge/orm/w0_orm_test.go.

## Acceptance (from the module root)
    grep -rn -i -E 'w0|TestW0' --include=*.go .        # no output
    gofmt -l api cli db log orm                         # empty
    go vet ./orm ./log ./api/errors ./cli/commands/migrations ./db/migrate/generate
    go test ./orm ./log ./api/errors ./cli/commands/migrations ./db/migrate/generate -count=1 -run 'NotImplemented|DefaultToPrimaryKey|SetNil' -v 2>&1 | grep -E '^(--- |ok|FAIL)'

The same 8 tests must pass under their new names. Report (max 15 lines): files created/deleted and each command result.
