package ahoy

import (
	"bytes"
	"testing"
)

func TestCLI_help(t *testing.T) {
	for _, tt := range []struct {
		name   string
		args   []string
		exit   int
		stdout string
		stderr string
	}{
		{
			"empty",
			nil,
			2,
			"",
			"ERROR: No command specified\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := CLI(tt.args, &stdout, &stderr)

			if exit != tt.exit {
				t.Errorf("CLI(%s) -> %d, want %d", tt.args, exit, tt.exit)
			}
		})
	}
}
