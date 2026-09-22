package cmd

import (
	"testing"
)

func TestCommandRegistration(t *testing.T) {
	cmds := rootCmd.Commands()
	expected := map[string]bool{
		"scan":     true,
		"query":    true,
		"explain":  true,
		"path":     true,
		"affected": true,
		"serve":    true,
	}

	for _, c := range cmds {
		delete(expected, c.Name())
	}

	if len(expected) > 0 {
		t.Errorf("missing commands in rootCmd: %v", expected)
	}
}
