package api

import (
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
)

type dummyAuth struct {
	tag string
}

func (d *dummyAuth) Authenticate(r *http.Request) (*authentication.AuthResult, error) {
	return nil, nil
}

func (d *dummyAuth) AuthenticateHeader(r *http.Request) string {
	return ""
}

type dummyPerm struct {
	tag string
}

func (p *dummyPerm) HasPermission(r *http.Request, view permissions.ViewSet) bool {
	return true
}

func (p *dummyPerm) HasObjectPermission(r *http.Request, view permissions.ViewSet, obj interface{}) bool {
	return true
}

func (p *dummyPerm) GetMessage() string {
	return ""
}

func (p *dummyPerm) GetCode() string {
	return ""
}

func TestSettings_Concurrency(t *testing.T) {
	orig := GetSettings()
	t.Cleanup(func() {
		SetSettings(orig)
	})

	const numWriters = 20
	const numReaders = 50

	var startWg sync.WaitGroup
	startWg.Add(1)

	var writersWg sync.WaitGroup
	writersWg.Add(numWriters)

	var stopReaders atomic.Bool
	var readersWg sync.WaitGroup
	readersWg.Add(numReaders)

	// 20 writer goroutines call SetDefaultAuthentication/SetDefaultPermissions/SetSettings
	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer writersWg.Done()
			startWg.Wait()
			for j := 0; j < 50; j++ {
				switch j % 3 {
				case 0:
					SetDefaultAuthentication(&dummyAuth{tag: "auth"})
				case 1:
					SetDefaultPermissions(&dummyPerm{tag: "perm"})
				case 2:
					s := DefaultSettings()
					s.PageSize = 20 + (j % 10)
					SetSettings(s)
				}
			}
		}(i)
	}

	// 50 reader goroutines call GetSettings().PageSize and read DefaultAuthentication len in a loop
	for i := 0; i < numReaders; i++ {
		go func() {
			defer readersWg.Done()
			startWg.Wait()
			for !stopReaders.Load() {
				s := GetSettings()
				_ = s.PageSize
				_ = len(s.DefaultAuthentication)
				_ = len(s.DefaultPermissions)
				runtime.Gosched()
			}
		}()
	}

	startWg.Done()
	writersWg.Wait()

	stopReaders.Store(true)
	readersWg.Wait()
}

func TestSettings_CopyOnWrite(t *testing.T) {
	orig := GetSettings()
	t.Cleanup(func() {
		SetSettings(orig)
	})

	p1 := &dummyPerm{tag: "p1"}
	SetDefaultPermissions(p1)

	before := GetSettings()
	if len(before.DefaultPermissions) != 1 || before.DefaultPermissions[0] != p1 {
		t.Fatalf("expected initial permission p1, got: %v", before.DefaultPermissions)
	}

	p2 := &dummyPerm{tag: "p2"}
	SetDefaultPermissions(p2)

	after := GetSettings()
	if before == after {
		t.Fatalf("expected different Settings pointers before and after SetDefaultPermissions (copy-on-write)")
	}

	// Pointer obtained BEFORE the call still has the old DefaultPermissions
	if len(before.DefaultPermissions) != 1 || before.DefaultPermissions[0] != p1 {
		t.Fatalf("Settings pointer obtained before call had its DefaultPermissions mutated; expected old DefaultPermissions preserved")
	}

	if len(after.DefaultPermissions) != 1 || after.DefaultPermissions[0] != p2 {
		t.Fatalf("Settings pointer obtained after call does not reflect new DefaultPermissions")
	}
}
