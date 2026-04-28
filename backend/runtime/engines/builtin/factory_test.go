package builtin

import (
	"testing"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/runtime/engines"
)

func TestNewRegistryFromConfigDetectsInstalledEngines(t *testing.T) {
	registry, err := NewRegistryFromConfig(&config.CLIEnginesConfig{})
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	if registry == nil {
		t.Fatal("expected registry")
	}
}

func TestNewEngineRejectsUnsupportedEngine(t *testing.T) {
	_, err := newEngine("unknown", "")
	if err == nil {
		t.Fatal("expected unsupported engine error")
	}
}

func TestNewEngineCreatesBuiltinEngines(t *testing.T) {
	for _, name := range []string{engines.EngineClaude, engines.EngineCodex} {
		engine, err := newEngine(name, name)
		if err != nil {
			t.Fatalf("build %s engine: %v", name, err)
		}
		if engine == nil {
			t.Fatalf("expected %s engine", name)
		}
	}
}
