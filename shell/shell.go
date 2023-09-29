package shell

import (
	"context"
	"io"
	"os"
	"os/exec"
)

const ShellToUse = "bash"

func Shellout(ctx context.Context, command string, additionalWriters ...io.Writer) error {
	writers := []io.Writer{os.Stdout, os.Stderr}
	writers = append(writers, additionalWriters...)
	multiWriter := io.MultiWriter(writers...)

	cmd := exec.CommandContext(ctx, ShellToUse, "-c", command)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter

	return cmd.Run()
}
