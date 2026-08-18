package imb

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrMailClass = errors.New("mail class denied")

// CopyIMBPrefix returns an independent prefix of the IMb payload bytes.
func CopyIMBPrefix(payload []byte, n int) []byte {
	if n < 0 {
		n = 0
	}
	if n > len(payload) {
		n = len(payload)
	}
	out := make([]byte, n)
	copy(out, payload[:n])
	return out
}

type TrayIndex struct {
	counts map[string]int
	order  []string
}

func NewTrayIndex() *TrayIndex {
	idx := &TrayIndex{}
	idx.counts = make(map[string]int)
	idx.order = make([]string, 0)
	return idx
}

func (t *TrayIndex) Add(tray string) {
	if _, ok := t.counts[tray]; !ok {
		t.order = append(t.order, tray)
	}
	t.counts[tray]++
}

func (t *TrayIndex) Count(tray string) int {
	return t.counts[tray]
}

func WrapClassDenied(op, class string) error {
	if strings.TrimSpace(op) == "" {
		op = "class-check"
	}
	if strings.TrimSpace(class) == "" {
		class = "unknown"
	}
	return fmt.Errorf("%s: class %s: %w", op, class, ErrMailClass)
}

func WaitInduction(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func DumpTrayList(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(body); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func GrowRouting(dst []byte, extra byte) []byte {
	out := make([]byte, len(dst)+1)
	copy(out, dst)
	out[len(dst)] = extra
	return out
}

type Piece struct {
	Barcode string
	Tray    string
}

func (p *Piece) Label() string {
	if p == nil {
		return ""
	}
	if p.Tray == "" {
		return p.Barcode
	}
	return p.Barcode + "@" + p.Tray
}

func ParsePieceMeta(b []byte) (map[string]string, error) {
	var m map[string]string
	if len(b) == 0 {
		return nil, errors.New("empty piece meta")
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func ExportInduction(root, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" {
		return "", errors.New("empty induction path")
	}
	if filepath.IsAbs(rel) {
		return "", errors.New("absolute induction path")
	}
	clean := filepath.Clean(rel)
	full := filepath.Join(root, clean)
	relOut, err := filepath.Rel(filepath.Clean(root), full)
	if err != nil {
		return "", err
	}
	if relOut == ".." || strings.HasPrefix(relOut, ".."+string(filepath.Separator)) {
		return "", errors.New("induction path escapes root")
	}
	return full, nil
}
