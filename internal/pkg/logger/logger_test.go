package logger

import "testing"

func TestNew_DebugAndInfo(t *testing.T) {
	l, err := New("debug")
	if err != nil || l == nil { t.Fatalf("debug logger: %v %v", l, err) }
	_ = l.Sync()
	l2, err := New("info")
	if err != nil || l2 == nil { t.Fatalf("info logger: %v %v", l2, err) }
	_ = l2.Sync()
}

func TestNew_InvalidLevelFallsBack(t *testing.T) {
	l, err := New("invalid-level")
	if err != nil || l == nil { t.Fatalf("logger should still be created: %v %v", l, err) }
	_ = l.Sync()
}
