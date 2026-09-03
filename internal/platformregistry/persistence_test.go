package platformregistry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersistentRegistryRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "platform-registry.json")
	registry, err := NewPersistent(path)
	if err != nil {
		t.Fatal(err)
	}
	record := fixture()
	if err := registry.Upsert(record); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("state mode = %o, want 600", info.Mode().Perm())
	}
	reloaded, err := NewPersistent(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reloaded.Get(record.Component.ID)
	if !ok || got.Source.Revision != record.Source.Revision {
		t.Fatalf("reloaded record = %#v, ok=%v", got, ok)
	}
}

func TestPersistentRegistryRejectsBroadPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "platform-registry.json")
	if err := os.WriteFile(path, []byte(`{"schema":"goreecloud.mesh.platform-registry-state.v1","records":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPersistent(path); err == nil {
		t.Fatal("expected broad state permissions to be rejected")
	}
}

func TestPersistentRegistryRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.json")
	if err := os.WriteFile(target, []byte(`{"schema":"goreecloud.mesh.platform-registry-state.v1","records":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "platform-registry.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPersistent(link); err == nil {
		t.Fatal("expected symlink state path to be rejected")
	}
}

func TestPersistentRegistryDoesNotCommitFailedRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "platform-registry.json")
	registry, err := NewPersistent(path)
	if err != nil {
		t.Fatal(err)
	}
	good := fixture()
	if err := registry.Upsert(good); err != nil {
		t.Fatal(err)
	}
	bad := fixture()
	bad.Source.AuthorityTransfer = true
	if err := registry.Upsert(bad); err == nil {
		t.Fatal("expected invalid record to fail")
	}
	got, ok := registry.Get(good.Component.ID)
	if !ok || got.Source.AuthorityTransfer {
		t.Fatal("failed record changed durable registry state")
	}
}
