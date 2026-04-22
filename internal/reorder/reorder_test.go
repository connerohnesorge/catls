package reorder

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/connerohnesorge/catls/internal/scanner"
)

func sampleFiles() []scanner.FileInfo {
	return []scanner.FileInfo{
		{Path: "/a", RelPath: "a.go"},
		{Path: "/b", RelPath: "b.go"},
		{Path: "/c", RelPath: "c.go"},
	}
}

func relPaths(files []scanner.FileInfo) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.RelPath
	}

	return out
}

func sendKey(t *testing.T, m *Model, keys string) {
	t.Helper()
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keys)})
}

func sendSpecial(t *testing.T, m *Model, key tea.KeyType) {
	t.Helper()
	_, _ = m.Update(tea.KeyMsg{Type: key})
}

func TestMoveItemDownAndUp_NoGrab(t *testing.T) {
	m := NewModel(sampleFiles())

	// shift+down on cursor 0: a <-> b, cursor lands on index 1
	sendKey(t, &m, "J")
	got := relPaths(m.Files())
	want := []string{"b.go", "a.go", "c.go"}
	if !equal(got, want) {
		t.Fatalf("after shift+down: got %v want %v", got, want)
	}
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}

	// shift+up: move back
	sendKey(t, &m, "K")
	got = relPaths(m.Files())
	want = []string{"a.go", "b.go", "c.go"}
	if !equal(got, want) {
		t.Fatalf("after shift+up: got %v want %v", got, want)
	}
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
}

func TestGrabMoveDrop(t *testing.T) {
	m := NewModel(sampleFiles())

	// grab
	sendKey(t, &m, " ")
	if !m.grabbed {
		t.Fatal("expected grabbed=true after space")
	}

	// while grabbed, down moves the item
	sendKey(t, &m, "j")
	sendKey(t, &m, "j")
	got := relPaths(m.Files())
	want := []string{"b.go", "c.go", "a.go"}
	if !equal(got, want) {
		t.Fatalf("after grab+jj: got %v want %v", got, want)
	}
	if m.cursor != 2 {
		t.Fatalf("cursor = %d, want 2", m.cursor)
	}

	// drop
	sendKey(t, &m, " ")
	if m.grabbed {
		t.Fatal("expected grabbed=false after space")
	}

	// with grab released, j only moves cursor (clamped)
	sendKey(t, &m, "j")
	got = relPaths(m.Files())
	if !equal(got, want) {
		t.Fatalf("cursor-only move changed list: got %v want %v", got, want)
	}
}

func TestCursorClamping(t *testing.T) {
	m := NewModel(sampleFiles())

	// up at top: stays at 0
	sendKey(t, &m, "k")
	if m.cursor != 0 {
		t.Fatalf("cursor at top after k = %d, want 0", m.cursor)
	}

	// move cursor to bottom
	sendKey(t, &m, "j")
	sendKey(t, &m, "j")
	sendKey(t, &m, "j") // beyond end — should clamp
	if m.cursor != 2 {
		t.Fatalf("cursor at bottom = %d, want 2", m.cursor)
	}
}

func TestMoveItemClamping(t *testing.T) {
	m := NewModel(sampleFiles())

	// shift+up at top: no-op
	sendKey(t, &m, "K")
	got := relPaths(m.Files())
	want := []string{"a.go", "b.go", "c.go"}
	if !equal(got, want) {
		t.Fatalf("shift+up at top mutated list: got %v", got)
	}

	// move cursor to bottom
	sendKey(t, &m, "j")
	sendKey(t, &m, "j")
	// shift+down at bottom: no-op
	sendKey(t, &m, "J")
	got = relPaths(m.Files())
	if !equal(got, want) {
		t.Fatalf("shift+down at bottom mutated list: got %v", got)
	}
}

func TestQuitDoesNotConfirm(t *testing.T) {
	m := NewModel(sampleFiles())
	sendKey(t, &m, "q")
	if !m.quitting {
		t.Fatal("expected quitting=true")
	}
	if m.confirmed {
		t.Fatal("expected confirmed=false on quit")
	}
}

func TestConfirm(t *testing.T) {
	m := NewModel(sampleFiles())
	sendSpecial(t, &m, tea.KeyEnter)
	if !m.confirmed {
		t.Fatal("expected confirmed=true after enter")
	}
	if m.quitting {
		t.Fatal("expected quitting=false after enter")
	}
}

func TestReorderEmpty(t *testing.T) {
	got, err := Reorder(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}

func TestFilesIsACopy(t *testing.T) {
	orig := sampleFiles()
	m := NewModel(orig)
	out := m.Files()
	out[0].RelPath = "mutated"
	if m.files[0].RelPath == "mutated" {
		t.Fatal("Files() must return a copy, not the backing slice")
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
