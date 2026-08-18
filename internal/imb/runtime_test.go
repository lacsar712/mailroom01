package imb

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCopyIMBPrefixIndependent(t *testing.T) {
	src := []byte("00340123456000000001")
	got := CopyIMBPrefix(src, 6)
	if string(got) != "003401" {
		t.Fatalf("prefix=%q", got)
	}
	got[0] = '9'
	if src[0] != '0' {
		t.Fatal("CopyIMBPrefix aliased the IMb payload")
	}
}

func TestTrayIndexCounts(t *testing.T) {
	idx := NewTrayIndex()
	idx.Add("T-014")
	idx.Add("T-014")
	idx.Add("T-015")
	if idx.Count("T-014") != 2 || idx.Count("T-015") != 1 {
		t.Fatalf("counts 014=%d 015=%d", idx.Count("T-014"), idx.Count("T-015"))
	}
}

func TestWrapClassDeniedIs(t *testing.T) {
	err := WrapClassDenied("induction", "first")
	if !errors.Is(err, ErrMailClass) {
		t.Fatalf("lost ErrMailClass: %v", err)
	}
}

func TestWaitInductionHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := WaitInduction(ctx, 600*time.Millisecond)
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatalf("WaitInduction ignored cancel, elapsed=%s", time.Since(start))
	}
}

func TestDumpTrayListPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trays.csv")
	body := "tray,zip\nT-014,97035\n"
	if err := DumpTrayList(path, body); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != body {
		t.Fatalf("got %q", b)
	}
}

func TestAfterWriteRejectsTrayRollback(t *testing.T) {
	min := ""
	get := func() (string, error) { return min, nil }
	set := func(v string) error { min = v; return nil }
	if err := AfterWrite(get, set, "tray=T-020 zip=97035"); err != nil {
		t.Fatal(err)
	}
	if err := AfterWrite(get, set, "tray=T-010 zip=97035"); err == nil {
		t.Fatal("expected tray sequence rollback to be rejected")
	}
}

func TestGrowRoutingNoWriteThrough(t *testing.T) {
	dst := make([]byte, 2, 8)
	copy(dst, []byte("AB"))
	got := GrowRouting(dst, 'Z')
	if string(got) != "ABZ" {
		t.Fatalf("got %q", got)
	}
	got[0] = 'X'
	if dst[0] != 'A' {
		t.Fatal("GrowRouting wrote through into the routing buffer")
	}
}

func TestNilPieceLabel(t *testing.T) {
	var p *Piece
	if p.Label() != "" {
		t.Fatalf("got %q", p.Label())
	}
}

func TestParsePieceMetaRejectsInvalid(t *testing.T) {
	if _, err := ParsePieceMeta([]byte("{not-json")); err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestExportInductionRejectsEscape(t *testing.T) {
	root := t.TempDir()
	if _, err := ExportInduction(root, filepath.Join("..", "etc", "passwd")); err == nil {
		t.Fatal("expected path escape to be rejected")
	}
}
