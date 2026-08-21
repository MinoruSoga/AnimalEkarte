//go:build !darwin

package labdeviceagent

import (
	"context"
	"errors"
	"io"
)

func openSerial(context.Context, string) (io.ReadCloser, error) {
	return nil, errors.New("lab device agent requires macOS")
}
