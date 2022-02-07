package infras

import (
	"context"
	"os/exec"
)

func ExecCommand(cmd string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*1e9)
	defer cancel()
	return exec.CommandContext(ctx, cmd, args...).CombinedOutput()
}
