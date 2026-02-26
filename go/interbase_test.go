package interbase

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasIC_WhenPresent(t *testing.T) {
	if _, err := exec.LookPath("ic"); err != nil {
		t.Skip("ic not on PATH")
	}
	if !HasIC() {
		t.Error("HasIC() = false, want true (ic is on PATH)")
	}
}

func TestHasIC_WhenMissing(t *testing.T) {
	old := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", old)

	if HasIC() {
		t.Error("HasIC() = true, want false (empty PATH)")
	}
}

func TestHasBD_WhenPresent(t *testing.T) {
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd not on PATH")
	}
	if !HasBD() {
		t.Error("HasBD() = false, want true (bd is on PATH)")
	}
}

func TestHasBD_WhenMissing(t *testing.T) {
	old := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", old)

	if HasBD() {
		t.Error("HasBD() = true, want false (empty PATH)")
	}
}

func TestHasCompanion_Empty(t *testing.T) {
	if HasCompanion("") {
		t.Error("HasCompanion('') = true, want false")
	}
}

func TestHasCompanion_Nonexistent(t *testing.T) {
	if HasCompanion("this-plugin-does-not-exist-zzzz") {
		t.Error("HasCompanion for nonexistent plugin should be false")
	}
}

func TestGetBead_Set(t *testing.T) {
	t.Setenv("CLAVAIN_BEAD_ID", "iv-test123")
	if got := GetBead(); got != "iv-test123" {
		t.Errorf("GetBead() = %q, want %q", got, "iv-test123")
	}
}

func TestGetBead_Unset(t *testing.T) {
	t.Setenv("CLAVAIN_BEAD_ID", "")
	if got := GetBead(); got != "" {
		t.Errorf("GetBead() = %q, want empty", got)
	}
}

func TestInEcosystem_FileExists(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "interbase.sh")
	os.WriteFile(path, []byte("#!/bin/bash"), 0644)

	t.Setenv("INTERMOD_LIB", path)
	if !InEcosystem() {
		t.Error("InEcosystem() = false, want true (file exists)")
	}
}

func TestInEcosystem_FileMissing(t *testing.T) {
	t.Setenv("INTERMOD_LIB", "/nonexistent/path/interbase.sh")
	if InEcosystem() {
		t.Error("InEcosystem() = true, want false (file does not exist)")
	}
}

func TestInSprint_NoBead(t *testing.T) {
	t.Setenv("CLAVAIN_BEAD_ID", "")
	if InSprint() {
		t.Error("InSprint() = true, want false (no bead)")
	}
}

func TestInSprint_NoIC(t *testing.T) {
	t.Setenv("CLAVAIN_BEAD_ID", "iv-test")
	old := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", old)

	if InSprint() {
		t.Error("InSprint() = true, want false (no ic)")
	}
}

func TestPhaseSet_NoBD(t *testing.T) {
	old := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", old)

	// Should succeed silently (fail-open) — no return value to check
	PhaseSet("bead-123", "planned")
}

func TestEmitEvent_NoIC(t *testing.T) {
	old := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", old)

	// Should succeed silently (fail-open) — no return value to check
	EmitEvent("run-123", "test-event")
}

func TestSessionStatus_Format(t *testing.T) {
	status := SessionStatus()
	if !strings.HasPrefix(status, "[interverse]") {
		t.Errorf("SessionStatus() = %q, want prefix [interverse]", status)
	}
	if !strings.Contains(status, "beads=") {
		t.Errorf("SessionStatus() = %q, should contain beads=", status)
	}
	if !strings.Contains(status, "ic=") {
		t.Errorf("SessionStatus() = %q, should contain ic=", status)
	}
}

func TestPluginCachePath_Empty(t *testing.T) {
	if got := PluginCachePath(""); got != "" {
		t.Errorf("PluginCachePath('') = %q, want empty", got)
	}
}

func TestEcosystemRoot_EnvOverride(t *testing.T) {
	t.Setenv("DEMARCH_ROOT", "/test/demarch")
	if got := EcosystemRoot(); got != "/test/demarch" {
		t.Errorf("EcosystemRoot() = %q, want /test/demarch", got)
	}
}

func TestEcosystemRoot_Unset(t *testing.T) {
	t.Setenv("DEMARCH_ROOT", "")
	// Should return something or empty — just shouldn't panic
	_ = EcosystemRoot()
}
