package migration

import (
	"log"
	"testing"
)

type nopWriter struct{}

func (w *nopWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func TestSetLogger(t *testing.T) {
	l := log.New(&nopWriter{}, "", 0)
	SetLogger(l)
	if logger != l {
		t.Errorf("SetLogger should set unexported logger to the given logger; expected %#v; logger is set to %#v\n", l, logger)
	}
}
