// Package interbase provides the Go SDK for Demarch plugin integration.
//
// All guard functions are fail-open: they return false when their dependency
// is missing. All action functions are silent no-ops when dependencies are
// absent. This ensures plugins work in both standalone and ecosystem modes.
package interbase

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// --- Guards ---

// HasIC returns true if the ic (Intercore) CLI is on PATH.
func HasIC() bool {
	_, err := exec.LookPath("ic")
	return err == nil
}

// HasBD returns true if the bd (Beads) CLI is on PATH.
func HasBD() bool {
	_, err := exec.LookPath("bd")
	return err == nil
}

// HasCompanion returns true if the named plugin is in the Claude Code cache.
func HasCompanion(name string) bool {
	if name == "" {
		return false
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	pattern := filepath.Join(home, ".claude", "plugins", "cache", "*", name, "*")
	matches, err := filepath.Glob(pattern)
	return err == nil && len(matches) > 0
}

// InEcosystem returns true if the centralized interbase install exists.
func InEcosystem() bool {
	path := os.Getenv("INTERMOD_LIB")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		path = filepath.Join(home, ".intermod", "interbase", "interbase.sh")
	}
	_, err := os.Stat(path)
	return err == nil
}

// GetBead returns the current bead ID from $CLAVAIN_BEAD_ID, or empty string.
func GetBead() string {
	return os.Getenv("CLAVAIN_BEAD_ID")
}

// InSprint returns true if there is an active sprint context (bead + ic run).
func InSprint() bool {
	if GetBead() == "" {
		return false
	}
	if !HasIC() {
		return false
	}
	cmd := exec.Command("ic", "run", "current", "--project=.")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// --- Actions ---
// Action functions return nothing — they are guaranteed no-ops when
// dependencies are absent. Returning error would create dead code at
// every call site (the spec says errors are never propagated).

// PhaseSet sets the phase on a bead. Silent no-op without bd.
func PhaseSet(bead, phase string, reason ...string) {
	if !HasBD() {
		return
	}
	cmd := exec.Command("bd", "set-state", bead, fmt.Sprintf("phase=%s", phase))
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[interbase] bd set-state failed: %v\n", err)
	}
}

// EmitEvent emits an event via ic. Silent no-op without ic.
func EmitEvent(runID, eventType string, payload ...string) {
	if !HasIC() {
		return
	}
	p := "{}"
	if len(payload) > 0 && payload[0] != "" {
		p = payload[0]
	}
	cmd := exec.Command("ic", "events", "emit", runID, eventType, fmt.Sprintf("--payload=%s", p))
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[interbase] ic events emit failed: %v\n", err)
	}
}

// SessionStatus returns the ecosystem status string.
func SessionStatus() string {
	var parts []string

	if HasBD() {
		parts = append(parts, "beads=active")
	} else {
		parts = append(parts, "beads=not-detected")
	}

	if HasIC() {
		cmd := exec.Command("ic", "run", "current", "--project=.")
		cmd.Stdout = nil
		cmd.Stderr = nil
		if cmd.Run() == nil {
			parts = append(parts, "ic=active")
		} else {
			parts = append(parts, "ic=not-initialized")
		}
	} else {
		parts = append(parts, "ic=not-detected")
	}

	return fmt.Sprintf("[interverse] %s", strings.Join(parts, " | "))
}

// --- Config + Discovery ---

// PluginCachePath returns the cache path for a named plugin.
// Returns empty string if not found.
func PluginCachePath(plugin string) string {
	if plugin == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	pattern := filepath.Join(home, ".claude", "plugins", "cache", "*", plugin, "*")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}
	// Return highest-versioned match (last after sort — Glob returns sorted)
	return matches[len(matches)-1]
}

// EcosystemRoot returns the Demarch monorepo root directory.
// Checks $DEMARCH_ROOT first, then walks up from CWD.
func EcosystemRoot() string {
	if root := os.Getenv("DEMARCH_ROOT"); root != "" {
		return root
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "sdk", "interbase")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// NudgeCompanion suggests installing a missing companion. Silent no-op if rate-limited.
// Rate-limited to 2 nudges per session with durable dismiss after 3 ignores.
func NudgeCompanion(companion, benefit string, plugin ...string) {
	if companion == "" {
		return
	}
	if HasCompanion(companion) {
		return
	}

	p := "unknown"
	if len(plugin) > 0 && plugin[0] != "" {
		p = plugin[0]
	}

	// Session budget check — sanitize session ID for safe filename
	sid := sanitizeID(os.Getenv("CLAUDE_SESSION_ID"))
	if sid == "" {
		sid = "unknown"
	}
	stateDir := filepath.Join(userConfigDir(), "interverse")
	sessionFile := filepath.Join(stateDir, fmt.Sprintf("nudge-session-%s.json", sid))

	count := readNudgeCount(sessionFile)
	if count >= 2 {
		return
	}

	// Durable dismissal check
	stateFile := filepath.Join(stateDir, "nudge-state.json")
	if isNudgeDismissed(stateFile, p, companion) {
		return
	}

	// Atomic dedup via mkdir — matches Bash/Python pattern. First caller wins.
	os.MkdirAll(stateDir, 0755)
	flag := filepath.Join(stateDir, fmt.Sprintf(".nudge-%s-%s-%s", sid, p, companion))
	if err := os.Mkdir(flag, 0755); err != nil {
		return // another hook already emitted this nudge
	}

	// Emit nudge
	fmt.Fprintf(os.Stderr, "[interverse] Tip: run /plugin install %s for %s.\n", companion, benefit)

	// Record
	writeNudgeCount(sessionFile, count+1)
	recordNudge(stateFile, p, companion)
}

// --- Internal helpers ---

// sanitizeID strips non-alphanumeric characters from an ID for safe filenames.
var safeIDRe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func sanitizeID(id string) string {
	return safeIDRe.ReplaceAllString(id, "")
}

func userConfigDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}

// nudgeSession is the JSON shape for session nudge budget files.
type nudgeSession struct {
	Count int `json:"count"`
}

// nudgeEntry is the JSON shape for per-companion nudge state.
type nudgeEntry struct {
	Ignores   int  `json:"ignores"`
	Dismissed bool `json:"dismissed"`
}

func readNudgeCount(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var s nudgeSession
	if err := json.Unmarshal(data, &s); err != nil {
		return 0
	}
	return s.Count
}

func writeNudgeCount(path string, count int) {
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.Marshal(nudgeSession{Count: count})
	os.WriteFile(path, data, 0644)
}

func isNudgeDismissed(stateFile, plugin, companion string) bool {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return false
	}
	var state map[string]nudgeEntry
	if err := json.Unmarshal(data, &state); err != nil {
		return false
	}
	key := plugin + ":" + companion
	entry, ok := state[key]
	return ok && entry.Dismissed
}

func recordNudge(stateFile, plugin, companion string) {
	os.MkdirAll(filepath.Dir(stateFile), 0755)
	key := plugin + ":" + companion

	var state map[string]nudgeEntry
	data, err := os.ReadFile(stateFile)
	if err != nil {
		state = make(map[string]nudgeEntry)
	} else {
		if err := json.Unmarshal(data, &state); err != nil {
			state = make(map[string]nudgeEntry)
		}
	}

	entry := state[key]
	entry.Ignores++
	if entry.Ignores >= 3 {
		entry.Dismissed = true
	}
	state[key] = entry

	out, _ := json.Marshal(state)
	os.WriteFile(stateFile, out, 0644)
}
