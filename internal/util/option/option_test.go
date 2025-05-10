package option_test

import (
	"testing"

	. "git.omicron.one/omicron/linkshare/internal/util/option"
)

func TestSome(t *testing.T) {
	opt := Some(42)

	if !opt.IsSome() {
		t.Error("Expected IsSome() to be true for Some(42)")
	}

	if opt.IsNone() {
		t.Error("Expected IsNone() to be false for Some(42)")
	}

	if opt.Value() != 42 {
		t.Errorf("Expected Value() to be 42, got %v", opt.Value())
	}

	if opt.ValueOr(0) != 42 {
		t.Errorf("Expected ValueOr(0) to be 42, got %v", opt.ValueOr(0))
	}
}

func TestNone(t *testing.T) {
	opt := None[int]()

	if opt.IsSome() {
		t.Error("Expected IsSome() to be false for None[int]()")
	}

	if !opt.IsNone() {
		t.Error("Expected IsNone() to be true for None[int]()")
	}

	if opt.ValueOr(99) != 99 {
		t.Errorf("Expected ValueOr(99) to be 99, got %v", opt.ValueOr(99))
	}
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Value() to panic on None")
		}
	}()

	opt := None[string]()
	_ = opt.Value() // This should panic
}
