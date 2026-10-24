package server

import (
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/stores/storestest"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductionRejectsAppDebug(t *testing.T) {
	settings := productionSettingsWith(secureTestSecurity())
	settings.App.Debug = true
	srv := &Server{settings: settings, config: &config.Config{Viper: viper.New()}}
	err := srv.validateProductionSecrets()
	if err == nil || !strings.Contains(err.Error(), "production requires app.debug=false") {
		t.Fatalf("validation error = %v", err)
	}
}

func TestProfilingRoutesRequireDebugAndEnableProfiling(t *testing.T) {
	for _, tc := range []struct{ debug, profiling, allowed bool }{{false, false, false}, {true, false, false}, {false, true, false}, {true, true, true}} {
		settings := &config.Settings{App: config.AppSettings{Env: "development", Debug: tc.debug}, Server: config.ServerSettings{EnableProfiling: tc.profiling, HealthCheckPath: "/health"}}
		srv, err := NewServer(&config.Config{Viper: viper.New()}, settings, nil, WithDatabase(storestest.OpenMemory(t)))
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		srv.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
		got := response.Code == http.StatusOK
		if got != tc.allowed {
			t.Fatalf("debug=%v profiling=%v status=%d", tc.debug, tc.profiling, response.Code)
		}
		if tc.allowed {
			unrelated := httptest.NewRecorder()
			srv.Handler.ServeHTTP(unrelated, httptest.NewRequest(http.MethodGet, "/debugger", nil))
			if unrelated.Code != http.StatusNotFound {
				t.Fatalf("profiling captured unrelated path: status=%d", unrelated.Code)
			}
		}
	}
}
