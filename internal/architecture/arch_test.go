package architecture_test

import (
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestArchitectureBoundaries(t *testing.T) {
	t.Helper()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedModule | packages.NeedFiles,
		Dir:  filepath.Clean(filepath.Join("..", "..")),
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		t.Fatalf("package load errors")
	}

	// Rules are intentionally minimal and high-signal.
	// Add rules incrementally as boundaries solidify.
	for _, p := range pkgs {
		// Rule 1: platform app assembly must not depend on model directly.
		if p.PkgPath == "easymail/internal/platform/app" {
			if _, ok := p.Imports["easymail/internal/model"]; ok {
				t.Fatalf("boundary violation: %s must not import easymail/internal/model", p.PkgPath)
			}
		}

		// Rule 2: model must not import redis client directly (cache belongs in higher layers).
		if strings.HasPrefix(p.PkgPath, "easymail/internal/model") {
			if _, ok := p.Imports["github.com/redis/go-redis/v9"]; ok {
				t.Fatalf("boundary violation: %s must not import github.com/redis/go-redis/v9", p.PkgPath)
			}
		}

		// Rule 3: services must not talk to database package directly; bootstrap should inject dependencies.
		// (Service packages should depend on ports/interfaces, not infrastructure singletons.)
		if strings.HasPrefix(p.PkgPath, "easymail/internal/service/") {
			if _, ok := p.Imports["easymail/internal/database"]; ok {
				t.Fatalf("boundary violation: %s must not import easymail/internal/database", p.PkgPath)
			}
		}

		// Rule 4: services must not depend on platform bootstrapping packages.
		if strings.HasPrefix(p.PkgPath, "easymail/internal/service/") {
			if _, ok := p.Imports["easymail/internal/platform/bootstrap"]; ok {
				t.Fatalf("boundary violation: %s must not import easymail/internal/platform/bootstrap", p.PkgPath)
			}
			if _, ok := p.Imports["easymail/internal/platform/app"]; ok {
				t.Fatalf("boundary violation: %s must not import easymail/internal/platform/app", p.PkgPath)
			}
			if _, ok := p.Imports["easymail/internal/platform/config"]; ok {
				t.Fatalf("boundary violation: %s must not import easymail/internal/platform/config", p.PkgPath)
			}
			if _, ok := p.Imports["easymail/internal/platform/runtime"]; ok {
				t.Fatalf("boundary violation: %s must not import easymail/internal/platform/runtime", p.PkgPath)
			}
		}
	}
}

