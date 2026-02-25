package toolerror_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/mistakeknot/interbase/toolerror"
)

func TestNew(t *testing.T) {
	te := toolerror.New(toolerror.ErrNotFound, "agent %q not found", "fd-safety")
	if te.Type != toolerror.ErrNotFound {
		t.Errorf("Type = %q, want %q", te.Type, toolerror.ErrNotFound)
	}
	if te.Message != `agent "fd-safety" not found` {
		t.Errorf("Message = %q, want %q", te.Message, `agent "fd-safety" not found`)
	}
	if te.Recoverable {
		t.Error("Recoverable should be false for NOT_FOUND")
	}
}

func TestTransientDefaultsRecoverable(t *testing.T) {
	te := toolerror.New(toolerror.ErrTransient, "database busy")
	if !te.Recoverable {
		t.Error("Recoverable should be true for TRANSIENT")
	}
}

func TestNonTransientDefaultsNotRecoverable(t *testing.T) {
	for _, typ := range []string{
		toolerror.ErrNotFound,
		toolerror.ErrConflict,
		toolerror.ErrValidation,
		toolerror.ErrPermission,
		toolerror.ErrInternal,
	} {
		te := toolerror.New(typ, "test")
		if te.Recoverable {
			t.Errorf("Type %q should default to Recoverable=false", typ)
		}
	}
}

func TestWithRecoverable(t *testing.T) {
	te := toolerror.New(toolerror.ErrConflict, "version mismatch").WithRecoverable(true)
	if !te.Recoverable {
		t.Error("WithRecoverable(true) should set Recoverable=true")
	}
}

func TestWithData(t *testing.T) {
	data := map[string]any{"file": "main.go", "line": 42}
	te := toolerror.New(toolerror.ErrValidation, "bad input").WithData(data)
	if te.Data["file"] != "main.go" {
		t.Errorf("Data[file] = %v, want main.go", te.Data["file"])
	}
}

func TestError(t *testing.T) {
	te := toolerror.New(toolerror.ErrPermission, "access denied")
	want := "[PERMISSION] access denied"
	if te.Error() != want {
		t.Errorf("Error() = %q, want %q", te.Error(), want)
	}
}

func TestJSON(t *testing.T) {
	te := toolerror.New(toolerror.ErrNotFound, "not found")
	var parsed map[string]any
	if err := json.Unmarshal([]byte(te.JSON()), &parsed); err != nil {
		t.Fatalf("JSON() produced invalid JSON: %v", err)
	}
	if parsed["type"] != toolerror.ErrNotFound {
		t.Errorf("type = %v, want %v", parsed["type"], toolerror.ErrNotFound)
	}
}

func TestFromError(t *testing.T) {
	te := toolerror.New(toolerror.ErrConflict, "conflict")
	// Direct ToolError
	got := toolerror.FromError(te)
	if got == nil {
		t.Fatal("FromError should unwrap a *ToolError")
	}
	if got.Type != toolerror.ErrConflict {
		t.Errorf("Type = %q, want %q", got.Type, toolerror.ErrConflict)
	}

	// Wrapped ToolError
	wrapped := fmt.Errorf("outer: %w", te)
	got = toolerror.FromError(wrapped)
	if got == nil {
		t.Fatal("FromError should unwrap through fmt.Errorf wrapping")
	}

	// Non-ToolError
	plain := errors.New("plain error")
	got = toolerror.FromError(plain)
	if got != nil {
		t.Error("FromError should return nil for non-ToolError")
	}
}

func TestWrap(t *testing.T) {
	// ToolError passes through
	te := toolerror.New(toolerror.ErrNotFound, "not found")
	got := toolerror.Wrap(te)
	if got.Type != toolerror.ErrNotFound {
		t.Errorf("Wrap(ToolError) should pass through, got Type=%q", got.Type)
	}

	// Plain error becomes ErrInternal
	plain := errors.New("something broke")
	got = toolerror.Wrap(plain)
	if got.Type != toolerror.ErrInternal {
		t.Errorf("Wrap(plain) Type = %q, want %q", got.Type, toolerror.ErrInternal)
	}
	if got.Message != "something broke" {
		t.Errorf("Wrap(plain) Message = %q, want %q", got.Message, "something broke")
	}
}
