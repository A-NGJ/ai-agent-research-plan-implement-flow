package search

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectionName(t *testing.T) {
	t.Run("same path yields same name", func(t *testing.T) {
		dir := t.TempDir()
		rpi := filepath.Join(dir, "myproject", ".rpi")
		n1, err := CollectionName(rpi)
		if err != nil {
			t.Fatal(err)
		}
		n2, err := CollectionName(rpi)
		if err != nil {
			t.Fatal(err)
		}
		if n1 != n2 {
			t.Errorf("expected stable name, got %q vs %q", n1, n2)
		}
	})

	t.Run("different paths yield different names (cross-project isolation)", func(t *testing.T) {
		root := t.TempDir()
		// Same final repo name but different absolute paths — short hash
		// must diverge so collections don't collide in qmd's global registry.
		a := filepath.Join(root, "alpha", "myproject", ".rpi")
		b := filepath.Join(root, "beta", "myproject", ".rpi")
		na, err := CollectionName(a)
		if err != nil {
			t.Fatal(err)
		}
		nb, err := CollectionName(b)
		if err != nil {
			t.Fatal(err)
		}
		if na == nb {
			t.Errorf("expected different names for different paths, both got %q", na)
		}
	})

	t.Run("slug handles uppercase, spaces, and special chars", func(t *testing.T) {
		cases := []struct {
			in   string
			want string
		}{
			{"My Project", "my-project"},
			{"Hello_World!!", "hello-world"},
			{"  spaced  ", "spaced"},
			{"already-clean", "already-clean"},
			{"___", ""},
			{"", ""},
		}
		for _, c := range cases {
			got := slugify(c.in)
			if got != c.want {
				t.Errorf("slugify(%q): got %q, want %q", c.in, got, c.want)
			}
		}
	})

	t.Run("slug capped at max length", func(t *testing.T) {
		long := strings.Repeat("x", 100)
		got := slugify(long)
		if len(got) > maxSlugLen {
			t.Errorf("expected slug capped at %d, got len %d", maxSlugLen, len(got))
		}
	})

	t.Run("collection name uses 'project' fallback when slug empty", func(t *testing.T) {
		dir := t.TempDir()
		// Create a path whose parent base slugifies to empty (special chars only).
		weirdParent := filepath.Join(dir, "___")
		rpi := filepath.Join(weirdParent, ".rpi")
		name, err := CollectionName(rpi)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(name, "rpi-project-") {
			t.Errorf("expected fallback prefix 'rpi-project-', got %q", name)
		}
	})

	t.Run("collection name has expected shape", func(t *testing.T) {
		dir := t.TempDir()
		rpi := filepath.Join(dir, "myrepo", ".rpi")
		name, err := CollectionName(rpi)
		if err != nil {
			t.Fatal(err)
		}
		// Shape: rpi-<slug>-<6 hex>
		if !strings.HasPrefix(name, "rpi-myrepo-") {
			t.Errorf("expected prefix 'rpi-myrepo-', got %q", name)
		}
		// Hash portion length: shortHashLen
		hash := strings.TrimPrefix(name, "rpi-myrepo-")
		if len(hash) != shortHashLen {
			t.Errorf("expected %d-char hash suffix, got %q (len %d)", shortHashLen, hash, len(hash))
		}
	})
}
