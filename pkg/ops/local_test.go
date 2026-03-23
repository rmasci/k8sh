package ops

import (
	"testing"
)

func TestGrepLocal(t *testing.T) {
	content := `line 1: hello world
line 2: Hello World
line 3: test pattern
line 4: another test
line 5: HELLO WORLD`

	t.Run("CaseSensitiveMatch", func(t *testing.T) {
		result := GrepLocal(content, "hello", false)
		
		expected := "line 1: hello world\n"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("CaseInsensitiveMatch", func(t *testing.T) {
		result := GrepLocal(content, "hello", true)
		
		expected := "line 1: hello world\nline 2: Hello World\nline 5: HELLO WORLD"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("NoMatch", func(t *testing.T) {
		result := GrepLocal(content, "notfound", false)
		
		if result != "" {
			t.Errorf("Expected empty result, got '%s'", result)
		}
	})
	
	t.Run("RegexPattern", func(t *testing.T) {
		result := GrepLocal(content, "line [0-9]", false)
		
		expected := `line 1: hello world
line 2: Hello World
line 3: test pattern
line 4: another test
line 5: HELLO WORLD`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("InvalidRegex", func(t *testing.T) {
		result := GrepLocal(content, "[invalid", false)
		
		if !contains(result, "Invalid regex") {
			t.Errorf("Expected invalid regex error, got '%s'", result)
		}
	})
}

func TestWordCountLocal(t *testing.T) {
	t.Run("MultipleLines", func(t *testing.T) {
		content := "line 1\nline 2\nline 3"
		result := WordCountLocal(content)
		
		expected := "3 6 17" // 3 lines, 6 words, 17 characters
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("EmptyContent", func(t *testing.T) {
		content := ""
		result := WordCountLocal(content)
		
		expected := "1 0 0" // 1 line, 0 words, 0 characters
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("SingleWord", func(t *testing.T) {
		content := "hello"
		result := WordCountLocal(content)
		
		expected := "1 1 5" // 1 line, 1 word, 5 characters
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("MultipleSpaces", func(t *testing.T) {
		content := "word1  word2   word3"
		result := WordCountLocal(content)
		
		expected := "1 3 21" // 1 line, 3 words, 21 characters
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

func TestSortLocal(t *testing.T) {
	t.Run("BasicSort", func(t *testing.T) {
		content := `zebra
apple
banana
cherry`
		result := SortLocal(content, false)
		
		expected := `apple
banana
cherry
zebra`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("UniqueSort", func(t *testing.T) {
		content := `apple
banana
apple
cherry
banana`
		result := SortLocal(content, true)
		
		expected := `apple
banana
cherry`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("EmptyLines", func(t *testing.T) {
		content := `apple

banana

cherry`
		result := SortLocal(content, false)
		
		expected := `apple
banana
cherry`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("EmptyContent", func(t *testing.T) {
		content := ""
		result := SortLocal(content, false)
		
		if result != "" {
			t.Errorf("Expected empty result, got '%s'", result)
		}
	})
	
	t.Run("SingleLine", func(t *testing.T) {
		content := "single line"
		result := SortLocal(content, false)
		
		if result != "single line" {
			t.Errorf("Expected 'single line', got '%s'", result)
		}
	})
}

func TestHeadLocal(t *testing.T) {
	content := `line 1
line 2
line 3
line 4
line 5`
	
	t.Run("DefaultLines", func(t *testing.T) {
		result := HeadLocal(content, 0)
		
		if result != content {
			t.Errorf("Expected full content, got '%s'", result)
		}
	})
	
	t.Run("SpecificLines", func(t *testing.T) {
		result := HeadLocal(content, 3)
		
		expected := `line 1
line 2
line 3`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("MoreLinesThanContent", func(t *testing.T) {
		result := HeadLocal(content, 10)
		
		if result != content {
			t.Errorf("Expected full content, got '%s'", result)
		}
	})
	
	t.Run("NegativeLines", func(t *testing.T) {
		result := HeadLocal(content, -1)
		
		if result != content {
			t.Errorf("Expected full content, got '%s'", result)
		}
	})
	
	t.Run("EmptyContent", func(t *testing.T) {
		result := HeadLocal("", 5)
		
		if result != "" {
			t.Errorf("Expected empty result, got '%s'", result)
		}
	})
}

func TestTailLocal(t *testing.T) {
	content := `line 1
line 2
line 3
line 4
line 5`
	
	t.Run("DefaultLines", func(t *testing.T) {
		result := TailLocal(content, 0)
		
		if result != content {
			t.Errorf("Expected full content, got '%s'", result)
		}
	})
	
	t.Run("SpecificLines", func(t *testing.T) {
		result := TailLocal(content, 3)
		
		expected := `line 3
line 4
line 5`
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
	
	t.Run("MoreLinesThanContent", func(t *testing.T) {
		result := TailLocal(content, 10)
		
		if result != content {
			t.Errorf("Expected full content, got '%s'", result)
		}
	})
	
	t.Run("NegativeLines", func(t *testing.T) {
		result := TailLocal(content, -1)
		
		if result != content {
			t.Errorf("Expected full content, got '%s'", result)
		}
	})
	
	t.Run("EmptyContent", func(t *testing.T) {
		result := TailLocal("", 5)
		
		if result != "" {
			t.Errorf("Expected empty result, got '%s'", result)
		}
	})
	
	t.Run("SingleLine", func(t *testing.T) {
		content := "single line"
		result := TailLocal(content, 5)
		
		if result != "single line" {
			t.Errorf("Expected 'single line', got '%s'", result)
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
