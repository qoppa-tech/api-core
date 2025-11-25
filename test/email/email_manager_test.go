package email

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/parlorhub/api-core/internal/email"
)

// setupTestTemplates creates a temporary directory with test templates
func setupTestTemplates(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	templates := map[string]struct {
		text string
		html string
	}{
		"welcome": {
			text: "Hello {{.Name}}!\nWelcome to our service.",
			html: "<html><body><h1>Hello {{.Name}}!</h1><p>Welcome to our service.</p></body></html>",
		},
		"password-reset": {
			text: "Hi {{.Name}},\nClick here to reset: {{.URL}}",
			html: "<html><body><p>Hi {{.Name}},</p><a href=\"{{.URL}}\">Reset Password</a></body></html>",
		},
		"newsletter": {
			text: "Newsletter for {{.Month}}\n{{.Content}}",
			html: "<html><body><h2>Newsletter for {{.Month}}</h2><div>{{.Content}}</div></body></html>",
		},
	}

	for name, content := range templates {
		dir := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create template dir: %v", err)
		}

		textPath := filepath.Join(dir, name+".txt")
		if err := os.WriteFile(textPath, []byte(content.text), 0o644); err != nil {
			t.Fatalf("failed to write text template: %v", err)
		}

		htmlPath := filepath.Join(dir, name+".html")
		if err := os.WriteFile(htmlPath, []byte(content.html), 0o644); err != nil {
			t.Fatalf("failed to write HTML template: %v", err)
		}
	}

	return tmpDir
}

func TestNewEmailManager(t *testing.T) {
	em := email.NewEmailManager("/test/path")

	if em.EmailTemplateFolder != "/test/path" {
		t.Errorf("expected folder /test/path, got %s", em.EmailTemplateFolder)
	}

	if em.TextCache == nil {
		t.Error("textCache should be initialized")
	}

	if em.HTMLCache == nil {
		t.Error("htmlCache should be initialized")
	}
}

func TestGenerateTemplate_Success(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	data := map[string]string{
		"Name": "Alice",
	}

	text, html, err := em.GenerateTemplate("welcome", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedText := "Hello Alice!\nWelcome to our service."
	if text != expectedText {
		t.Errorf("expected text %q, got %q", expectedText, text)
	}

	if !strings.Contains(html, "Hello Alice!") {
		t.Errorf("HTML should contain 'Hello Alice!', got %q", html)
	}

	if !strings.Contains(html, "<html>") {
		t.Errorf("HTML should contain HTML tags, got %q", html)
	}
}

func TestGenerateTemplate_WithURL(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	data := map[string]string{
		"Name": "Bob",
		"URL":  "https://example.com/reset?token=abc123",
	}

	text, html, err := em.GenerateTemplate("password-reset", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(text, "Bob") {
		t.Errorf("text should contain 'Bob', got %q", text)
	}

	if !strings.Contains(text, "https://example.com/reset?token=abc123") {
		t.Errorf("text should contain URL, got %q", text)
	}

	if !strings.Contains(html, "Bob") {
		t.Errorf("HTML should contain 'Bob', got %q", html)
	}

	if !strings.Contains(html, "https://example.com/reset?token=abc123") {
		t.Errorf("HTML should contain URL, got %q", html)
	}
}

func TestGenerateTemplate_TemplateNotFound(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	data := map[string]string{"Name": "Charlie"}

	_, _, err := em.GenerateTemplate("nonexistent", data)
	if err == nil {
		t.Error("expected error for nonexistent template, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load templates") {
		t.Errorf("error should mention failed to load templates, got %v", err)
	}
}

func TestGenerateTemplate_MissingTextTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "incomplete")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	htmlPath := filepath.Join(dir, "incomplete.html")
	if err := os.WriteFile(htmlPath, []byte("<html>{{.Name}}</html>"), 0o644); err != nil {
		t.Fatalf("failed to write HTML: %v", err)
	}

	em := email.NewEmailManager(tmpDir)
	data := map[string]string{"Name": "Dave"}

	_, _, err := em.GenerateTemplate("incomplete", data)
	if err == nil {
		t.Error("expected error for missing text template, got nil")
	}
}

func TestGenerateTemplate_InvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "invalid")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	textPath := filepath.Join(dir, "invalid.txt")
	if err := os.WriteFile(textPath, []byte("{{.Name"), 0o644); err != nil {
		t.Fatalf("failed to write text: %v", err)
	}

	htmlPath := filepath.Join(dir, "invalid.html")
	if err := os.WriteFile(htmlPath, []byte("<html>{{.Name}}</html>"), 0o644); err != nil {
		t.Fatalf("failed to write HTML: %v", err)
	}

	em := email.NewEmailManager(tmpDir)
	data := map[string]string{"Name": "Eve"}

	_, _, err := em.GenerateTemplate("invalid", data)
	if err == nil {
		t.Error("expected error for invalid template syntax, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("error should mention parsing failure, got %v", err)
	}
}

func TestGenerateTemplate_MissingProperty(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	// Missing required properties
	data := map[string]string{}

	text, html, err := em.GenerateTemplate("welcome", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Templates should still execute but with empty values
	if !strings.Contains(text, "Hello <no value>!") {
		t.Errorf("text should contain 'Hello <no value>!', got %q", text)
	}

	if !strings.Contains(html, "Hello !") {
		t.Errorf("HTML should contain 'Hello <no value>!', got %q", html)
	}
}

func TestTemplateCache(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	data := map[string]string{"Name": "Frank"}

	// First call - templates should be parsed and cached
	_, _, err := em.GenerateTemplate("welcome", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(em.TextCache) != 1 {
		t.Errorf("expected 1 cached text template, got %d", len(em.TextCache))
	}

	if len(em.HTMLCache) != 1 {
		t.Errorf("expected 1 cached HTML template, got %d", len(em.HTMLCache))
	}

	// Second call - should use cached templates
	_, _, err = em.GenerateTemplate("welcome", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cache size should remain the same
	if len(em.TextCache) != 1 {
		t.Errorf("expected 1 cached text template after second call, got %d", len(em.TextCache))
	}

	if len(em.HTMLCache) != 1 {
		t.Errorf("expected 1 cached HTML template after second call, got %d", len(em.HTMLCache))
	}
}

func TestClearCache(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	data := map[string]string{"Name": "Grace"}

	_, _, err := em.GenerateTemplate("welcome", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(em.TextCache) == 0 || len(em.HTMLCache) == 0 {
		t.Error("cache should not be empty after generating template")
	}

	em.ClearCache()

	if len(em.TextCache) != 0 {
		t.Errorf("expected empty text cache after clear, got %d", len(em.TextCache))
	}

	if len(em.HTMLCache) != 0 {
		t.Errorf("expected empty HTML cache after clear, got %d", len(em.HTMLCache))
	}
}

func TestPreloadTemplates_Success(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	err := em.PreloadTemplates("welcome", "password-reset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(em.TextCache) != 2 {
		t.Errorf("expected 2 cached text templates, got %d", len(em.TextCache))
	}

	if len(em.HTMLCache) != 2 {
		t.Errorf("expected 2 cached HTML templates, got %d", len(em.HTMLCache))
	}

	// Verify we can use the preloaded templates
	data := map[string]string{"Name": "Henry"}
	_, _, err = em.GenerateTemplate("welcome", data)
	if err != nil {
		t.Fatalf("unexpected error using preloaded template: %v", err)
	}
}

func TestPreloadTemplates_NonexistentTemplate(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	err := em.PreloadTemplates("welcome", "nonexistent")
	if err == nil {
		t.Error("expected error when preloading nonexistent template, got nil")
	}

	if !strings.Contains(err.Error(), "failed to preload template") {
		t.Errorf("error should mention preload failure, got %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Simulate concurrent template generation
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			data := map[string]string{
				"Name": "User",
			}

			_, _, err := em.GenerateTemplate("welcome", data)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}

	// Verify cache is still consistent
	if len(em.TextCache) != 1 {
		t.Errorf("expected 1 cached text template after concurrent access, got %d", len(em.TextCache))
	}

	if len(em.HTMLCache) != 1 {
		t.Errorf("expected 1 cached HTML template after concurrent access, got %d", len(em.HTMLCache))
	}
}

func TestGenerateTemplate_WithStruct(t *testing.T) {
	tmpDir := setupTestTemplates(t)
	em := email.NewEmailManager(tmpDir)

	type User struct {
		Name string
	}

	data := User{Name: "Iris"}

	text, html, err := em.GenerateTemplate("welcome", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(text, "Iris") {
		t.Errorf("text should contain 'Iris', got %q", text)
	}

	if !strings.Contains(html, "Iris") {
		t.Errorf("HTML should contain 'Iris', got %q", html)
	}
}

func BenchmarkGenerateTemplate_Cached(b *testing.B) {
	tmpDir := setupTestTemplates(&testing.T{})
	em := email.NewEmailManager(tmpDir)

	data := map[string]string{"Name": "Benchmark"}

	// Warm up cache
	em.GenerateTemplate("welcome", data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := em.GenerateTemplate("welcome", data)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkGenerateTemplate_Uncached(b *testing.B) {
	tmpDir := setupTestTemplates(&testing.T{})
	data := map[string]string{"Name": "Benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em := email.NewEmailManager(tmpDir)
		_, _, err := em.GenerateTemplate("welcome", data)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
