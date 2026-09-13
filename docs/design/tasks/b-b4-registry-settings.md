TASK B4: synchronize the plugin registry and global API settings (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/registry-settings-concurrency/forge
(git worktree on branch fix/registry-settings-concurrency, based on master.)

Files you may modify — ONLY:
  registry/plugin.go
  api/api.go
  tests: NEW registry/plugin_concurrency_test.go, NEW api/settings_concurrency_test.go
Do NOT edit go.mod / go.sum. Do NOT touch any other file.

Write failing tests FIRST (they must fail or report DATA RACE under `go test -race`).

=== BUG 1 (High): PluginRegistry has no lock ===
registry/plugin.go:
    type PluginRegistry struct {
        plugins      map[string]Plugin
        modelPlugins []ModelPlugin
        adminPlugins []AdminPlugin
        apiPlugins   []APIPlugin
    }
RegisterPlugin reads and writes the map and appends to the slices with no synchronization;
GetPlugin / GetAllPlugins read the map; GetModelPlugins / GetAdminPlugins / GetAPIPlugins return
the INTERNAL slices, so callers can mutate registry state and race with appends.

Fix:
1. Add `mu sync.RWMutex` to PluginRegistry.
2. RegisterPlugin — must NOT hold the lock while calling plugin code (Install, Extend*, apply*),
   because plugin code may call GetPlugin and would deadlock:
     a. mu.Lock(); if name exists -> mu.Unlock(), return the existing "already registered" error;
        otherwise insert plugins[name] = plugin (reserves the name); mu.Unlock()
     b. err := plugin.Install(); on error: mu.Lock(); delete(plugins, name); mu.Unlock();
        return the existing wrapped error
     c. mu.Lock(); append to the typed slices the plugin satisfies; mu.Unlock()
     d. call applyModelExtensions / applyAdminExtensions / applyAPIExtensions exactly as today,
        with no lock held.
   Keep the existing error messages and ordering of Install vs extension application.
3. GetPlugin, GetAllPlugins: RLock. GetAllPlugins already copies the map — keep that.
4. GetModelPlugins, GetAdminPlugins, GetAPIPlugins: RLock and return a COPY of the slice
   (`append([]T(nil), s...)`).
5. If any other function in registry/plugin.go touches globalPluginRegistry fields, apply the same
   locking. Grep the file for `globalPluginRegistry.`.

Tests (registry/plugin_concurrency_test.go):
  - 50 goroutines each RegisterPlugin a uniquely named fake plugin while 50 goroutines call
    GetAllPlugins / GetModelPlugins / GetPlugin in a loop; -race clean; afterwards all 50 present.
  - Two goroutines register the SAME name concurrently: exactly one succeeds.
  - A plugin whose Install fails is not left registered (GetPlugin returns error) and its name can
    be registered again.
  - A plugin whose Install calls GetPlugin(its own name) does not deadlock (use a 2s timeout via a
    done channel + select).
  - Mutating the slice returned by GetModelPlugins does not change a second call's result.
  Tests share the global registry: give every plugin a unique name (t.Name() + index) and do not
  assume the registry starts empty. If the package has a reset helper for tests, use it.

=== BUG 2 (High): global API settings read and written without synchronization ===
api/api.go:
    var globalSettings = DefaultSettings()
    func SetSettings(settings *Settings) { globalSettings = settings }
    func GetSettings() *Settings { return globalSettings }
    func SetDefaultAuthentication(authClasses ...) { settings := GetSettings(); settings.DefaultAuthentication = authClasses }
    func SetDefaultPermissions(permClasses ...)    { settings := GetSettings(); settings.DefaultPermissions = permClasses }
The pointer swap is a data race with every request reading settings, and the setters mutate the
shared struct in place while requests read it.

Fix:
1. Replace the variable with `var globalSettings atomic.Pointer[Settings]` initialised in an
   `init()` (or a package-level func) to DefaultSettings(), plus `var settingsWriteMu sync.Mutex`.
2. SetSettings: settingsWriteMu.Lock(); globalSettings.Store(settings); Unlock.
   (Keep accepting the caller's pointer; document that callers must not mutate it afterwards.)
3. GetSettings: return globalSettings.Load().
4. SetDefaultAuthentication / SetDefaultPermissions (and ANY other function in api/api.go that
   mutates the struct returned by GetSettings — grep for `GetSettings()` in api/api.go):
   copy-on-write:
       settingsWriteMu.Lock(); defer settingsWriteMu.Unlock()
       cur := globalSettings.Load()
       next := *cur                          // shallow copy
       next.DefaultAuthentication = append([]authentication.Authentication(nil), authClasses...)
       globalSettings.Store(&next)
   Note Settings may contain a sync type or noCopy — check the struct; if it cannot be copied
   (go vet copylocks), report it and instead protect with a package RWMutex around read/write.
5. Initialize() keeps working (it builds settings then calls SetSettings).

Tests (api/settings_concurrency_test.go):
  - 20 goroutines call SetDefaultAuthentication/SetDefaultPermissions/SetSettings while 50 call
    GetSettings().PageSize and read DefaultAuthentication len in a loop; -race clean.
  - After SetDefaultPermissions(p1), a Settings pointer obtained BEFORE the call still has the old
    DefaultPermissions (copy-on-write proven).
  - Restore the original settings at test end with t.Cleanup(func(){ api.SetSettings(orig) }).

=== VERIFY ===
  gofmt -l registry/plugin.go api/api.go <your test files>   (prints nothing)
  go vet ./registry/... ./api/...
  go test -race -count=2 ./registry/ ./api/
  go test ./...   (TestPrefetchRelated_Integration in ./orm is known-flaky)

Also, REPORT ONLY (do not change): list any other package-level registries in forge/registry/*.go
(extensions.go, registry.go, type_registry.go) that mutate maps without a mutex, with file:line.

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No exported signature changes. Minimal diff.
- Final report: files changed, tests added, the report-only list, exact tail of the -race run.
