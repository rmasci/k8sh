package editor

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewViEditor(t *testing.T) {
	content := "test content\nline 2"
	filename := "test.txt"
	
	editor := NewViEditor(content, filename)
	
	if editor.filename != filename {
		t.Errorf("Expected filename '%s', got '%s'", filename, editor.filename)
	}
	
	if editor.mode != NormalMode {
		t.Errorf("Expected mode to be NormalMode, got %v", editor.mode)
	}
	
	if editor.cursor.X != 0 || editor.cursor.Y != 0 {
		t.Errorf("Expected cursor at (0,0), got (%d,%d)", editor.cursor.X, editor.cursor.Y)
	}
	
	if editor.modified {
		t.Error("Expected modified to be false initially")
	}
	
	if len(editor.lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(editor.lines))
	}
	
	if editor.lines[0] != "test content" {
		t.Errorf("Expected first line 'test content', got '%s'", editor.lines[0])
	}
	
	if editor.lines[1] != "line 2" {
		t.Errorf("Expected second line 'line 2', got '%s'", editor.lines[1])
	}
}

func TestNewViEditorEmptyContent(t *testing.T) {
	content := ""
	filename := "test.txt"
	
	editor := NewViEditor(content, filename)
	
	if len(editor.lines) != 1 {
		t.Errorf("Expected 1 line for empty content, got %d", len(editor.lines))
	}
	
	if editor.lines[0] != "" {
		t.Errorf("Expected empty line, got '%s'", editor.lines[0])
	}
}

func TestViEditorGetters(t *testing.T) {
	content := "line 1\nline 2"
	filename := "test.txt"
	editor := NewViEditor(content, filename)
	
	t.Run("GetContent", func(t *testing.T) {
		result := editor.GetContent()
		if result != content {
			t.Errorf("Expected '%s', got '%s'", content, result)
		}
	})
	
	t.Run("IsModified", func(t *testing.T) {
		if editor.IsModified() {
			t.Error("Expected IsModified to be false initially")
		}
		
		editor.modified = true
		if !editor.IsModified() {
			t.Error("Expected IsModified to be true after modification")
		}
	})
	
	t.Run("GetMode", func(t *testing.T) {
		if editor.GetMode() != NormalMode {
			t.Errorf("Expected NormalMode, got %v", editor.GetMode())
		}
		
		editor.mode = InsertMode
		if editor.GetMode() != InsertMode {
			t.Errorf("Expected InsertMode, got %v", editor.GetMode())
		}
	})
}

func TestViEditorSetMode(t *testing.T) {
	editor := NewViEditor("test", "test.txt")
	
	editor.SetMode(InsertMode)
	if editor.mode != InsertMode {
		t.Errorf("Expected InsertMode, got %v", editor.mode)
	}
	
	editor.SetMode(VisualMode)
	if editor.mode != VisualMode {
		t.Errorf("Expected VisualMode, got %v", editor.mode)
	}
	
	// Setting mode to non-visual should disable selection
	editor.selection.Active = true
	editor.SetMode(NormalMode)
	if editor.selection.Active {
		t.Error("Expected selection to be inactive when leaving VisualMode")
	}
}

func TestViEditorMoveCursor(t *testing.T) {
	editor := NewViEditor("line 1\nline 2\nline 3", "test.txt")
	
	t.Run("BasicMovement", func(t *testing.T) {
		editor.MoveCursor(1, 0) // Right
		if editor.cursor.X != 1 || editor.cursor.Y != 0 {
			t.Errorf("Expected cursor at (1,0), got (%d,%d)", editor.cursor.X, editor.cursor.Y)
		}
		
		editor.MoveCursor(0, 1) // Down
		if editor.cursor.X != 1 || editor.cursor.Y != 1 {
			t.Errorf("Expected cursor at (1,1), got (%d,%d)", editor.cursor.X, editor.cursor.Y)
		}
		
		editor.MoveCursor(-1, 0) // Left
		if editor.cursor.X != 0 || editor.cursor.Y != 1 {
			t.Errorf("Expected cursor at (0,1), got (%d,%d)", editor.cursor.X, editor.cursor.Y)
		}
		
		editor.MoveCursor(0, -1) // Up
		if editor.cursor.X != 0 || editor.cursor.Y != 0 {
			t.Errorf("Expected cursor at (0,0), got (%d,%d)", editor.cursor.X, editor.cursor.Y)
		}
	})
	
	t.Run("BoundaryChecking", func(t *testing.T) {
		// Try to move beyond top
		editor.MoveCursor(0, -10)
		if editor.cursor.Y < 0 {
			t.Errorf("Y should not be negative, got %d", editor.cursor.Y)
		}
		
		// Try to move beyond bottom
		editor.cursor.Y = 2
		editor.MoveCursor(0, 10)
		if editor.cursor.Y >= len(editor.lines) {
			t.Errorf("Y should not exceed line count, got %d", editor.cursor.Y)
		}
		
		// Try to move beyond left
		editor.MoveCursor(-10, 0)
		if editor.cursor.X < 0 {
			t.Errorf("X should not be negative, got %d", editor.cursor.X)
		}
		
		// Try to move beyond right
		editor.cursor.X = 3
		editor.MoveCursor(10, 0)
		if editor.cursor.X > len(editor.lines[editor.cursor.Y]) {
			t.Errorf("X should not exceed line length, got %d", editor.cursor.X)
		}
	})
}

func TestViEditorInsertChar(t *testing.T) {
	editor := NewViEditor("hello", "test.txt")
	editor.SetMode(InsertMode)
	
	editor.InsertChar('X')
	
	if editor.lines[0] != "Xhello" {
		t.Errorf("Expected 'Xhello', got '%s'", editor.lines[0])
	}
	
	if editor.cursor.X != 1 {
		t.Errorf("Expected cursor X to be 1, got %d", editor.cursor.X)
	}
	
	if !editor.modified {
		t.Error("Expected modified to be true after insertion")
	}
}

func TestViEditorDeleteChar(t *testing.T) {
	editor := NewViEditor("hello", "test.txt")
	editor.cursor.X = 1
	
	editor.DeleteChar()
	
	if editor.lines[0] != "hllo" {
		t.Errorf("Expected 'hllo', got '%s'", editor.lines[0])
	}
	
	if !editor.modified {
		t.Error("Expected modified to be true after deletion")
	}
}

func TestViEditorDeleteLine(t *testing.T) {
	editor := NewViEditor("line 1\nline 2\nline 3", "test.txt")
	editor.cursor.Y = 1
	
	editor.DeleteLine()
	
	if len(editor.lines) != 2 {
		t.Errorf("Expected 2 lines after deletion, got %d", len(editor.lines))
	}
	
	if editor.lines[0] != "line 1" {
		t.Errorf("Expected first line 'line 1', got '%s'", editor.lines[0])
	}
	
	if editor.lines[1] != "line 3" {
		t.Errorf("Expected second line 'line 3', got '%s'", editor.lines[1])
	}
	
	if editor.cursor.Y != 1 {
		t.Errorf("Expected cursor Y to be 1, got %d", editor.cursor.Y)
	}
	
	if editor.cursor.X != 0 {
		t.Errorf("Expected cursor X to be 0, got %d", editor.cursor.X)
	}
}

func TestViEditorDeleteLineSingleLine(t *testing.T) {
	editor := NewViEditor("single line", "test.txt")
	
	editor.DeleteLine()
	
	if len(editor.lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(editor.lines))
	}
	
	if editor.lines[0] != "" {
		t.Errorf("Expected empty line, got '%s'", editor.lines[0])
	}
	
	if editor.cursor.X != 0 {
		t.Errorf("Expected cursor X to be 0, got %d", editor.cursor.X)
	}
}

func TestViEditorYankLine(t *testing.T) {
	editor := NewViEditor("line 1\nline 2", "test.txt")
	editor.cursor.Y = 1
	
	editor.YankLine()
	
	if editor.clipboard.Content != "line 2" {
		t.Errorf("Expected clipboard content 'line 2', got '%s'", editor.clipboard.Content)
	}
	
	if editor.clipboard.Type != LineClipboard {
		t.Errorf("Expected clipboard type LineClipboard, got %v", editor.clipboard.Type)
	}
}

func TestViEditorPaste(t *testing.T) {
	editor := NewViEditor("line 1\nline 3", "test.txt")
	editor.cursor.Y = 0
	editor.clipboard.Content = "line 2"
	editor.clipboard.Type = LineClipboard
	
	t.Run("PasteAfter", func(t *testing.T) {
		editor.Paste(true)
		
		if len(editor.lines) != 3 {
			t.Errorf("Expected 3 lines after paste, got %d", len(editor.lines))
		}
		
		if editor.lines[1] != "line 2" {
			t.Errorf("Expected second line 'line 2', got '%s'", editor.lines[1])
		}
		
		if editor.cursor.Y != 1 {
			t.Errorf("Expected cursor Y to be 1, got %d", editor.cursor.Y)
		}
	})
	
	t.Run("PasteBefore", func(t *testing.T) {
		editor.cursor.Y = 0
		editor.Paste(false)
		
		if len(editor.lines) != 4 {
			t.Errorf("Expected 4 lines after second paste, got %d", len(editor.lines))
		}
		
		if editor.lines[0] != "line 2" {
			t.Errorf("Expected first line 'line 2', got '%s'", editor.lines[0])
		}
	})
}

func TestViEditorSearch(t *testing.T) {
	editor := NewViEditor("hello world\nhello universe\ngoodbye world", "test.txt")
	
	t.Run("BasicSearch", func(t *testing.T) {
		editor.Search("hello", true)
		
		if len(editor.searchMatches) != 2 {
			t.Errorf("Expected 2 matches, got %d", len(editor.searchMatches))
		}
		
		if editor.currentMatch != 0 {
			t.Errorf("Expected current match to be 0, got %d", editor.currentMatch)
		}
		
		if editor.cursor.Y != 0 || editor.cursor.X != 0 {
			t.Errorf("Expected cursor at first match (0,0), got (%d,%d)", editor.cursor.X, editor.cursor.Y)
		}
	})
	
	t.Run("SearchBackward", func(t *testing.T) {
		editor.Search("hello", false)
		
		if editor.currentMatch != 1 {
			t.Errorf("Expected current match to be 1 (last match), got %d", editor.currentMatch)
		}
	})
	
	t.Run("EmptySearch", func(t *testing.T) {
		editor.Search("", true)
		
		if len(editor.searchMatches) != 0 {
			t.Errorf("Expected no matches for empty search, got %d", len(editor.searchMatches))
		}
	})
}

func TestViEditorNextMatch(t *testing.T) {
	editor := NewViEditor("hello world\nhello world\nhello world", "test.txt")
	editor.Search("hello", true)
	
	initialMatch := editor.currentMatch
	editor.NextMatch()
	
	if editor.currentMatch == initialMatch {
		t.Error("Expected current match to change")
	}
	
	// Should wrap around
	for i := 0; i < 3; i++ {
		editor.NextMatch()
	}
	
	if editor.currentMatch != initialMatch {
		t.Errorf("Expected to wrap around to initial match %d, got %d", initialMatch, editor.currentMatch)
	}
}

func TestViEditorPrevMatch(t *testing.T) {
	editor := NewViEditor("hello world\nhello world\nhello world", "test.txt")
	editor.Search("hello", true)
	
	initialMatch := editor.currentMatch
	editor.PrevMatch()
	
	if editor.currentMatch == initialMatch {
		t.Error("Expected current match to change")
	}
	
	// Should wrap around
	for i := 0; i < 3; i++ {
		editor.PrevMatch()
	}
	
	if editor.currentMatch != initialMatch {
		t.Errorf("Expected to wrap around to initial match %d, got %d", initialMatch, editor.currentMatch)
	}
}

func TestViEditorExecuteCommand(t *testing.T) {
	editor := NewViEditor("test content", "test.txt")
	
	t.Run("WriteCommand", func(t *testing.T) {
		result := editor.ExecuteCommand("w")
		if result != "save" {
			t.Errorf("Expected 'save', got '%s'", result)
		}
	})
	
	t.Run("WriteWithFilename", func(t *testing.T) {
		result := editor.ExecuteCommand("w newfile.txt")
		if result != "save" {
			t.Errorf("Expected 'save', got '%s'", result)
		}
		if editor.filename != "newfile.txt" {
			t.Errorf("Expected filename to be 'newfile.txt', got '%s'", editor.filename)
		}
	})
	
	t.Run("QuitCommand", func(t *testing.T) {
		result := editor.ExecuteCommand("q")
		if result != "quit" {
			t.Errorf("Expected 'quit', got '%s'", result)
		}
	})
	
	t.Run("QuitWithChanges", func(t *testing.T) {
		editor.modified = true
		result := editor.ExecuteCommand("q")
		if result != "No write since last change" {
			t.Errorf("Expected 'No write since last change', got '%s'", result)
		}
	})
	
	t.Run("WriteQuitCommand", func(t *testing.T) {
		result := editor.ExecuteCommand("wq")
		if result != "save_quit" {
			t.Errorf("Expected 'save_quit', got '%s'", result)
		}
	})
	
	t.Run("SubstituteCommand", func(t *testing.T) {
		result := editor.ExecuteCommand("s/old/new")
		if result != "Pattern not found" {
			t.Errorf("Expected 'Pattern not found', got '%s'", result)
		}
	})
	
	t.Run("SubstituteGlobal", func(t *testing.T) {
		editor.lines[0] = "old old old"
		result := editor.ExecuteCommand("s/old/new/g")
		if !contains(result, "Replaced") {
			t.Errorf("Expected replacement message, got '%s'", result)
		}
		if editor.lines[0] != "new new new" {
			t.Errorf("Expected 'new new new', got '%s'", editor.lines[0])
		}
	})
}

func TestViModel(t *testing.T) {
	editor := NewViEditor("test", "test.txt")
	model := ViModel{editor: editor, quit: false}
	
	t.Run("Init", func(t *testing.T) {
		cmd := model.Init()
		if cmd != nil {
			t.Error("Expected nil command from Init")
		}
	})
	
	t.Run("UpdateWindowSize", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		newModel, _ := model.Update(msg)
		
		viModel := newModel.(ViModel)
		if viModel.editor.width != 100 {
			t.Errorf("Expected width 100, got %d", viModel.editor.width)
		}
		if viModel.editor.height != 50 {
			t.Errorf("Expected height 50, got %d", viModel.editor.height)
		}
	})
}

func TestAllStringIndices(t *testing.T) {
	t.Run("MultipleMatches", func(t *testing.T) {
		indices := allStringIndices("hello world hello", "hello")
		expected := []int{0, 12}
		
		if len(indices) != len(expected) {
			t.Errorf("Expected %d indices, got %d", len(expected), len(indices))
		}
		
		for i, idx := range expected {
			if indices[i] != idx {
				t.Errorf("Expected index %d at position %d, got %d", idx, i, indices[i])
			}
		}
	})
	
	t.Run("SingleMatch", func(t *testing.T) {
		indices := allStringIndices("hello world", "hello")
		expected := []int{0}
		
		if len(indices) != 1 || indices[0] != 0 {
			t.Errorf("Expected [0], got %v", indices)
		}
	})
	
	t.Run("NoMatch", func(t *testing.T) {
		indices := allStringIndices("hello world", "xyz")
		
		if len(indices) != 0 {
			t.Errorf("Expected no indices, got %v", indices)
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}
