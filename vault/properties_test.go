package vault

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempVaultWithFile(t *testing.T, name, content string) (*Client, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name+".md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	c := New(dir)
	if err := c.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return c, path
}

func TestUpdatePageProperties_NoFrontmatter_WritesFrontmatter(t *testing.T) {
	body := "This is the note body.\n\nSome content here.\n"
	c, path := tempVaultWithFile(t, "my-note", body)
	ctx := context.Background()

	err := c.UpdatePageProperties(ctx, "my-note", map[string]any{
		"type":   "meeting",
		"status": "archived",
	})
	if err != nil {
		t.Fatalf("UpdatePageProperties: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	result := string(got)

	if !strings.HasPrefix(result, "---\n") {
		t.Errorf("expected YAML frontmatter, got:\n%s", result)
	}
	if strings.Contains(result, "type::") || strings.Contains(result, "status::") {
		t.Errorf("found inline Logseq-style properties, expected YAML frontmatter:\n%s", result)
	}
	if !strings.Contains(result, "type: meeting") {
		t.Errorf("missing 'type: meeting' in frontmatter:\n%s", result)
	}
	if !strings.Contains(result, "status: archived") {
		t.Errorf("missing 'status: archived' in frontmatter:\n%s", result)
	}
	if !strings.Contains(result, "This is the note body.") {
		t.Errorf("body content was lost:\n%s", result)
	}
}

func TestUpdatePageProperties_ExistingFrontmatter_MergesKeys(t *testing.T) {
	initial := "---\ndomain: UKG\n---\nExisting body.\n"
	c, path := tempVaultWithFile(t, "existing-note", initial)
	ctx := context.Background()

	err := c.UpdatePageProperties(ctx, "existing-note", map[string]any{
		"status": "archived",
	})
	if err != nil {
		t.Fatalf("UpdatePageProperties: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	result := string(got)

	if !strings.Contains(result, "domain: UKG") {
		t.Errorf("original key 'domain' lost:\n%s", result)
	}
	if !strings.Contains(result, "status: archived") {
		t.Errorf("new key 'status' not written:\n%s", result)
	}
	if !strings.Contains(result, "Existing body.") {
		t.Errorf("body content was lost:\n%s", result)
	}
}

func TestUpdatePageProperties_OverwritesExistingKey(t *testing.T) {
	initial := "---\nstatus: active\n---\nBody.\n"
	c, path := tempVaultWithFile(t, "overwrite-note", initial)
	ctx := context.Background()

	err := c.UpdatePageProperties(ctx, "overwrite-note", map[string]any{
		"status": "archived",
	})
	if err != nil {
		t.Fatalf("UpdatePageProperties: %v", err)
	}

	got, _ := os.ReadFile(path)
	result := string(got)

	if strings.Contains(result, "status: active") {
		t.Errorf("old value 'active' still present:\n%s", result)
	}
	if !strings.Contains(result, "status: archived") {
		t.Errorf("new value 'archived' not written:\n%s", result)
	}
}

func TestUpdatePageProperties_PageNotFound_ReturnsError(t *testing.T) {
	c, _ := tempVaultWithFile(t, "real-note", "body\n")
	err := c.UpdatePageProperties(context.Background(), "nonexistent", map[string]any{"k": "v"})
	if err == nil {
		t.Fatal("expected error for missing page, got nil")
	}
}
