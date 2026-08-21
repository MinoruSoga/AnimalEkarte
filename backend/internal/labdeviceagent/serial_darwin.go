//go:build darwin

package labdeviceagent

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func openSerial(ctx context.Context, path string) (io.ReadCloser, error) {
	if !strings.HasPrefix(path, "/dev/cu.usbserial-") {
		return nil, fmt.Errorf("unsupported serial path")
	}
	command := exec.CommandContext(ctx, "/bin/stty", "-f", path, "9600", "cs8", "-cstopb", "-parenb", "raw", "-echo")
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("configure serial port: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open serial port: %w", err)
	}
	return file, nil
}
