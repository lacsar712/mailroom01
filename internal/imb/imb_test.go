package imb

import "testing"

func TestAccept(t *testing.T) {
	if err := Enforce("00340123456000000001", "tray=T-1", []string{"first"}); err != nil {
		t.Fatal(err)
	}
}

func TestRejectTitle(t *testing.T) {
	if err := Enforce("hello", "tray=T-1", []string{"first"}); err == nil {
		t.Fatal("expected reject")
	}
}
