package imb

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Rec struct {
	Title, Body string
	Tags        []string
}

func Sample() Rec {
	return Rec{Title: "00340123456000000001", Body: "tray=T-014 zip=97035", Tags: []string{"first"}}
}

func Seed() []Rec {
	return []Rec{
		Sample(),
		{Title: "00340123456000000002", Body: "tray=T-014 zip=97035", Tags: []string{"standard"}},
	}
}

func AfterWrite(getMin func() (string, error), setMin func(string) error, body string) error {
	tray := parseTray(body)
	n := trayNumber(tray)
	if n < 0 {
		return fmt.Errorf("tray sequence missing in %q", body)
	}
	cur, err := getMin()
	if err == nil && strings.TrimSpace(cur) != "" {
		old, conv := strconv.Atoi(strings.TrimSpace(cur))
		if conv == nil && n < old {
			return fmt.Errorf("tray sequence %d < last committed %d", n, old)
		}
	}
	return setMin(strconv.Itoa(n))
}

func parseTray(body string) string {
	for _, part := range strings.Fields(body) {
		k, v, ok := strings.Cut(part, "=")
		if ok && k == "tray" {
			return v
		}
	}
	return ""
}

func trayNumber(tray string) int {
	n := 0
	found := false
	for _, r := range tray {
		if r >= '0' && r <= '9' {
			found = true
			n = n*10 + int(r-'0')
		}
	}
	if !found {
		return -1
	}
	return n
}

func Steps() []string { return []string{"barcode-check", "tray-index", "induction-export"} }

func Enforce(title, body string, tags []string) error {
	if !validIMB(title) {
		return fmt.Errorf("title must be 20-31 digit IMb payload")
	}
	if !strings.Contains(body, "tray=") {
		return fmt.Errorf("body must contain tray=")
	}
	if len(tags) == 0 {
		return fmt.Errorf("mail class tag required")
	}
	return nil
}

func validIMB(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 20 || len(s) > 31 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
