package sessions

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockTmuxRunner mocks tmux session output.
type MockTmuxRunner struct {
	sessions []TmuxSession
	err      error
}

func (m *MockTmuxRunner) ListSessions() ([]TmuxSession, error) {
	return m.sessions, m.err
}

// MockFileInfo mocks os.FileInfo.
type MockFileInfo struct {
	name  string
	isDir bool
}

func (m MockFileInfo) Name() string       { return m.name }
func (m MockFileInfo) Size() int64        { return 0 }
func (m MockFileInfo) Mode() os.FileMode  { return 0 }
func (m MockFileInfo) ModTime() time.Time { return time.Now() }
func (m MockFileInfo) IsDir() bool        { return m.isDir }
func (m MockFileInfo) Sys() interface{}   { return nil }

// MockDirEntry mocks os.DirEntry.
type MockDirEntry struct {
	name  string
	isDir bool
}

func (m MockDirEntry) Name() string               { return m.name }
func (m MockDirEntry) IsDir() bool               { return m.isDir }
func (m MockDirEntry) Type() fs.FileMode          { return 0 }
func (m MockDirEntry) Info() (os.FileInfo, error) { return MockFileInfo{name: m.name, isDir: m.isDir}, nil }

// MockFileSystem mocks the filesystem access.
type MockFileSystem struct {
	dirs  map[string][]os.DirEntry
	files map[string][]byte
	links map[string]string
	stats map[string]os.FileInfo
}

func (m *MockFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	if entries, ok := m.dirs[dirname]; ok {
		return entries, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	if data, ok := m.files[filename]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Readlink(name string) (string, error) {
	if target, ok := m.links[name]; ok {
		return target, nil
	}
	return "", os.ErrNotExist
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if info, ok := m.stats[name]; ok {
		return info, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if m.files == nil {
		m.files = make(map[string][]byte)
	}
	m.files[filename] = data
	if m.stats == nil {
		m.stats = make(map[string]os.FileInfo)
	}
	m.stats[filename] = MockFileInfo{name: filepath.Base(filename), isDir: false}
	return nil
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (m *MockFileSystem) Remove(name string) error {
	if m.files != nil {
		delete(m.files, name)
	}
	if m.stats != nil {
		delete(m.stats, name)
	}
	return nil
}

func TestClassifySessions(t *testing.T) {
	mockTmux := &MockTmuxRunner{
		sessions: []TmuxSession{
			{Name: "emagy-session-12345", Path: "/workspace/proj1", Attached: true, Windows: 2},
			{Name: "emgem-session-abcde", Path: "/workspace/proj2", Attached: false, Windows: 1},
			{Name: "my-private-session", Path: "/home/ricc", Attached: true, Windows: 3},
		},
	}

	mockFS := &MockFileSystem{
		files: map[string][]byte{
			"/home/ricc/.gemini/antigravity-cli/cache/last_conversations.json": []byte(`{
				"/workspace/proj1": "session-12345",
				"/workspace/proj3": "session-Chumbia",
				"/workspace/proj4": "session-chumbawamba",
				"/workspace/proj5": "session-active-but-dead"
			}`),
			"/home/ricc/.gemini/antigravity-cli/brain/session-active-but-dead/.system_generated/logs/transcript.jsonl": []byte(`{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","content":"<USER_REQUEST>\nImplement a new feature\n</USER_REQUEST>"}`),
		},
		dirs: map[string][]os.DirEntry{},
	}

	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	t.Run("Exclusion Rules (Default/Chumbia)", func(t *testing.T) {
		t.Setenv("EXCLUDE_RESUSCITATE", "")
		sessions, err := engine.Classify(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Let's verify statuses:
		// 1. emagy-session-12345 -> open_tmux (agy)
		// 2. emgem-session-abcde -> open_tmux (gemini)
		// 3. my-private-session -> open_private (unknown)
		// 4. session-Chumbia -> dead_archived (due to Chumbia)
		// 5. session-chumbawamba -> dead_archived (due to chumbawamba)
		// 6. session-active-but-dead -> dead_resuscitatable

		foundAgy := false
		foundGem := false
		foundPrivate := false
		foundChumbia := false
		foundChumba := false
		foundActiveButDead := false

		for _, s := range sessions {
			switch s.ID {
			case "session-12345", "emagy-session-12345":
				foundAgy = true
				if s.State != StateOpenTmux || s.Harness != "agy" {
					t.Errorf("expected session-12345 to be open_tmux/agy, got %s/%s", s.State, s.Harness)
				}
			case "emgem-session-abcde":
				foundGem = true
				if s.State != StateOpenTmux || s.Harness != "gemini" {
					t.Errorf("expected emgem-session-abcde to be open_tmux/gemini, got %s/%s", s.State, s.Harness)
				}
			case "my-private-session":
				foundPrivate = true
				if s.State != StateOpenPrivate {
					t.Errorf("expected my-private-session to be open_private, got %s", s.State)
				}
			case "session-Chumbia":
				foundChumbia = true
				if s.State != StateDeadArchived {
					t.Errorf("expected session-Chumbia to be dead_archived, got %s", s.State)
				}
			case "session-chumbawamba":
				foundChumba = true
				if s.State != StateDeadArchived {
					t.Errorf("expected session-chumbawamba to be dead_archived, got %s", s.State)
				}
			case "session-active-but-dead":
				foundActiveButDead = true
				if s.State != StateDeadResuscitatable {
					t.Errorf("expected session-active-but-dead to be dead_resuscitatable, got %s", s.State)
				}
				if s.Description != "Implement a new feature" {
					t.Errorf("expected description to be 'Implement a new feature', got %q", s.Description)
				}
			}
		}

		if !foundAgy {
			t.Error("missing emagy session")
		}
		if !foundGem {
			t.Error("missing emgem session")
		}
		if !foundPrivate {
			t.Error("missing private session")
		}
		if !foundChumbia {
			t.Error("missing Chumbia session")
		}
		if !foundChumba {
			t.Error("missing chumbawamba session")
		}
		if !foundActiveButDead {
			t.Error("missing dead resuscitatable session")
		}
	})

	t.Run("Exclusion Rules (Custom Env Excludes)", func(t *testing.T) {
		t.Setenv("EXCLUDE_RESUSCITATE", "*active*")
		sessions, err := engine.Classify(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		foundActiveButDead := false
		for _, s := range sessions {
			if s.ID == "session-active-but-dead" {
				foundActiveButDead = true
				if s.State != StateDeadArchived {
					t.Errorf("expected session-active-but-dead to be dead_archived via custom exclude, got %s", s.State)
				}
			}
		}
		if !foundActiveButDead {
			t.Error("missing session-active-but-dead")
		}
	})
}
