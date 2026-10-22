package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/forgego/forge/admin/components"
	"github.com/forgego/forge/admin/core"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type contractPlugin struct{ core.BasePlugin }

func (*contractPlugin) ID() string   { return "contract" }
func (*contractPlugin) Name() string { return "Contract" }
func (*contractPlugin) GetPages() map[string]components.Component {
	return map[string]components.Component{"overview": components.Page("Overview")}
}
func (*contractPlugin) GetMenuItems() []core.MenuItem {
	return []core.MenuItem{{Label: "Overview", Path: "/plugins/contract/pages/overview"}}
}

func TestDocumentedSiteRegistrationMountAndPluginContract(t *testing.T) {
	var registerDefault func(*Config[storeNote]) (*Admin[storeNote], error) = Register[storeNote]
	require.NotNil(t, registerDefault)
	site := NewSite("contract")
	_, hasAddWidget := reflect.TypeOf(site).MethodByName("AddWidget")
	require.False(t, hasAddWidget, "dashboard configuration is global rather than a Site widget method")
	require.Equal(t, "contract", site.Name)
	site.Title = "Contract Admin"
	ui := DefaultUIConfig()
	ui.Prefix = "/admin"
	site.WithUIConfig(ui)
	require.Equal(t, "Contract Admin", site.Title)
	require.Equal(t, "/admin", site.GetUIConfig().Prefix)
	registered, err := RegisterWithSite(site, &Config[storeNote]{
		ListDisplay:         []Field{Computed("title")},
		SearchFields:        []Field{Computed("title")},
		ListFilter:          []Field{Computed("title")},
		Ordering:            []Field{Computed("-title")},
		ReadOnlyFields:      []string{"title"},
		HasAddPermission:    func(context.Context, *Admin[storeNote], interface{}) bool { return true },
		HasViewPermission:   func(context.Context, *Admin[storeNote], interface{}, *storeNote) bool { return true },
		HasChangePermission: func(context.Context, *Admin[storeNote], interface{}, *storeNote) bool { return true },
		HasDeletePermission: func(context.Context, *Admin[storeNote], interface{}, *storeNote) bool { return true },
		Actions: []Action[storeNote]{{
			Name: "review", Label: "Review", Confirmation: "Review selected?",
			Permissions: []string{"review"},
			Handler:     func(context.Context, []*storeNote) error { return nil },
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "-title", registered.ConfigInterface().(*core.Config[storeNote]).Ordering[0].Path())
	require.NoError(t, site.RegisterPlugin(context.Background(), &contractPlugin{}))
	require.Error(t, site.RegisterPlugin(context.Background(), &contractPlugin{}), "duplicate plugin IDs must fail")
	plugin, err := site.GetRegistry().GetPlugin("contract")
	require.NoError(t, err)
	require.Equal(t, components.TypePage, plugin.GetPages()["overview"].Type)
	r := chi.NewRouter()
	r.Mount("/admin", site.Handler())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/", nil))
	require.Equal(t, http.StatusOK, rec.Code)
}
