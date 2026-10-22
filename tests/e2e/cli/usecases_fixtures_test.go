package cli

// This is developer-authored integration code, not a replacement for generated
// models or endpoints. Each real application receives its own expanded copy.
const useCaseRoutes = `package work

import (
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io"
    "net/http"

    "github.com/forgego/forge/admin"
    admincore "github.com/forgego/forge/admin/core"
    "github.com/forgego/forge/api"
    "github.com/forgego/forge/api/authentication"
    "github.com/forgego/forge/api/permissions"
    "github.com/forgego/forge/api/throttling"
    "github.com/forgego/forge/db"
    "github.com/forgego/forge/server"
)

// ownerPolicy adds the application's create/transfer rule to Forge's object
// permission. IsOwnerOrReadOnly alone deliberately permits public reads and
// does not assign an owner on create.
type ownerPolicy struct{}
func (ownerPolicy) HasPermission(r *http.Request, _ permissions.ViewSet) bool {
    if permissions.IsSafeMethod(r.Method) || r.Body == nil { return true }
    user, ok := authentication.GetUserFromRequest(r)
    member, valid := user.(*Member)
    if !ok || !valid { return false }
    raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
    if err != nil { return false }
    r.Body = io.NopCloser(bytes.NewReader(raw))
    var body map[string]json.RawMessage
    if json.Unmarshal(raw, &body) != nil { return true } // framework returns 400
    // The API accepts schema/JSON/database and Go field names. Inspect all
    // accepted owner aliases so the application rule cannot be bypassed.
    for _, key := range []string{"owner_id", "OwnerId"} {
        if value, exists := body[key]; exists {
            var owner int64
            if json.Unmarshal(value, &owner) != nil || owner != member.Id { return false }
        }
    }
    return true
}
func (ownerPolicy) HasObjectPermission(r *http.Request, v permissions.ViewSet, obj interface{}) bool {
    return permissions.NewIsOwnerOrReadOnly("owner_id").HasObjectPermission(r, v, obj)
}
func (ownerPolicy) GetMessage() string { return "only the owner may change this record" }
func (ownerPolicy) GetCode() string { return "permission_denied" }

func memberForToken(token string) (interface{}, error) {
    digest := sha256.Sum256([]byte(token))
    qs, err := MemberObjects.Filter(MemberFieldsInstance.ApiToken.Eq(hex.EncodeToString(digest[:])))
    if err != nil { return nil, err }
    rows, err := qs.Limit(1).All(context.Background())
    if err != nil || len(rows) == 0 { return nil, err }
    return rows[0], nil
}

func Register(router *server.Router, database *db.DB) {
    {{range .Models}}{{.Name}}Objects.SetDB(database)
    {{end}}
    api.SetDefaultAuthentication(authentication.NewTokenAuthentication(memberForToken))
    api.SetDefaultPermissions(permissions.NewIsAuthenticated())
    apiRouter := api.NewRouter("/api/v1")
    {{range .Models}}
    view{{.Name}} := New{{.Name}}ViewSet()
    {{if .ReadOnly}}view{{.Name}}.ReadOnly = true{{end}}
    {{if .Owned}}view{{.Name}}.Permissions = []permissions.Permission{permissions.NewIsAuthenticated(), ownerPolicy{}}{{end}}
    {{if .StaffOnly}}view{{.Name}}.Permissions = []permissions.Permission{permissions.NewIsAuthenticated(), permissions.NewIsStaffUser()}{{end}}
    apiRouter.Register("{{.Route}}", view{{.Name}})
    {{end}}
    limited := New{{.Primary}}ViewSet()
    limited.ReadOnly = true
    limited.Throttles = []throttling.Throttle{throttling.NewUserRateThrottle("2/min")}
    apiRouter.Register("limited", limited)
    apiRouter.RegisterRoutes(router)

    {{range .Models}}
    {{if .Primary}}_, err{{.Name}} := admin.Register[{{.Name}}](&admincore.Config[{{.Name}}]{
        ListDisplay: []admincore.Field{ {{.Name}}FieldsInstance.Id, {{.Name}}FieldsInstance.Name },
        SearchFields: []admincore.Field{ {{.Name}}FieldsInstance.Name },
        ListPerPage: 2,
    })
    {{else}}_, err{{.Name}} := admin.Register[{{.Name}}](&admincore.Config[{{.Name}}]{})
    {{end}}
    if err{{.Name}} != nil { panic(err{{.Name}}) }
    {{end}}
}
`

// An application consumer test uses the generated manager API and real
// PostgreSQL, including a failing transaction and independently checked counts.
const useCaseORMTest = `package work

import (
    "context"
    "errors"
    "fmt"
    "testing"

    "github.com/forgego/forge/config"
    "github.com/forgego/forge/db"
    "github.com/forgego/forge/orm"
)

func TestConsumerManagerContract(t *testing.T) {
    database, err := db.NewDBFromConfig(config.NewConfig())
    if err != nil { t.Fatal(err) }
    defer database.Close()
    {{.Primary}}Objects.SetDB(database)
    ctx := context.Background()
    before, err := {{.Primary}}Objects.Count(ctx)
    if err != nil { t.Fatal(err) }
    sentinel := errors.New("business operation aborted")
    err = database.WithTx(ctx, func(tx *db.Tx) error {
        item, err := {{.Primary}}Objects.New()
        if err != nil { return err }
        item.Name = "rolled-back-consumer"
        {{.UniqueAssign}}
        if err := {{.Primary}}Objects.WithTx(tx).Create(ctx, item); err != nil { return err }
        return sentinel
    })
    if !errors.Is(err, sentinel) { t.Fatalf("transaction: %v", err) }
    after, err := {{.Primary}}Objects.Count(ctx)
    if err != nil || after != before { t.Fatalf("rollback changed count: %d -> %d (%v)", before, after, err) }
    qs := {{.Primary}}Objects.QuerySet()
    values, err := orm.AggregateValues(ctx, qs, orm.Count("id"))
    if err != nil { t.Fatal(err) }
    if fmt.Sprint(values["count"]) != fmt.Sprint(before) { t.Fatalf("aggregate count: %v, want %d", values, before) }
    rows, err := qs.OrderBy("name").Limit(1).All(ctx)
    if err != nil || len(rows) != 1 { t.Fatalf("typed queryset: %v, %v", rows, err) }
    found, err := {{.Primary}}Objects.Get(ctx, rows[0].Id)
    if err != nil || found.Name != rows[0].Name { t.Fatalf("manager Get: %v, %v", found, err) }
    found.Name = "consumer-updated"
    if err := {{.Primary}}Objects.Update(ctx, found); err != nil { t.Fatal(err) }
    // Restore the fixture for subsequent HTTP/browser journeys.
    found.Name = rows[0].Name
    if err := {{.Primary}}Objects.Save(ctx, found); err != nil { t.Fatal(err) }
}
`
