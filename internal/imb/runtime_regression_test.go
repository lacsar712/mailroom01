package imb

import (
	"bytes"
	"testing"
)

// TestCopyIMBPrefixNoWriteThrough ensures the returned prefix is an independent
// copy: writing into it must never reach back into the inbound barcode source
// buffer. Regression for the aliasing bug where CopyIMBPrefix returned a reslice.
func TestCopyIMBPrefixNoWriteThrough(t *testing.T) {
	src := []byte("00340123456000000001")
	want := []byte("003401")

	got := CopyIMBPrefix(src, 6)
	if !bytes.Equal(got, want) {
		t.Fatalf("prefix=%q want=%q", got, want)
	}

	// Mutating the returned prefix must not affect the source buffer.
	got[0] = '9'
	got[len(got)-1] = '9'
	if src[0] != '0' || src[5] != '1' {
		t.Fatalf("source buffer polluted by prefix write: src=%q", src)
	}
}

// TestCopyIMBPrefixBounds covers the length-clamping contract that callers rely
// on so a stray large n cannot read past the payload.
func TestCopyIMBPrefixBounds(t *testing.T) {
	src := []byte("00340123456000000001")

	if got := CopyIMBPrefix(src, -1); len(got) != 0 {
		t.Fatalf("negative n should clamp to 0, got %d bytes", len(got))
	}
	if got := CopyIMBPrefix(src, 0); len(got) != 0 {
		t.Fatalf("n=0 should yield empty slice, got %d bytes", len(got))
	}
	if got := CopyIMBPrefix(src, len(src)); !bytes.Equal(got, src) {
		t.Fatalf("n=len should copy whole payload, got %q", got)
	}
	// n past the end must clamp to len(src) and still be an independent copy.
	got := CopyIMBPrefix(src, len(src)+5)
	if len(got) != len(src) {
		t.Fatalf("oversize n should clamp to len, got %d bytes", len(got))
	}
	got[0] = 'X'
	if src[0] != '0' {
		t.Fatal("oversize-n copy still aliased the source buffer")
	}
}

// TestCopyIMBPrefixNilEmpty guards the zero-value inputs used when an inbound
// barcode has not yet been read.
func TestCopyIMBPrefixNilEmpty(t *testing.T) {
	if got := CopyIMBPrefix(nil, 4); len(got) != 0 {
		t.Fatalf("nil payload should yield empty slice, got %d bytes", len(got))
	}
	if got := CopyIMBPrefix([]byte{}, 4); len(got) != 0 {
		t.Fatalf("empty payload should yield empty slice, got %d bytes", len(got))
	}
}
