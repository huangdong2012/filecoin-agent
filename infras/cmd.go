package infras

import (
	"context"
	"os/exec"
)

func ExecCommand(cmd string, args ...string) ([]byte, error) {
	return exec.CommandContext(context.Background(), cmd, args...).CombinedOutput()
}
