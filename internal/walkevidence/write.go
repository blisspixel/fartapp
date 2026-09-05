package walkevidence

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteNew publishes a synced, complete archive without replacing any existing
// destination. Hard-link publication is atomic and must be supported by the
// chosen filesystem. No archive-supplied path is used. Cancellation observed
// before publication leaves no destination. Cancellation racing with the link
// operation can still return a successful publication; a committed destination
// is never removed. Abrupt process death can leave a private temporary file;
// this API does not promise directory durability after power loss.
func WriteNew(ctx context.Context, destination string, record Record) error {
	if ctx == nil || len(record.encoded) == 0 {
		return fmt.Errorf("%w: missing context or record", ErrInvalidArchive)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	directory, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer directory.Close()
	name := ".fartevidence-" + rand.Text() + ".tmp"
	file, err := directory.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer directory.Remove(name)
	if err := stage(ctx, file, record.encoded); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := directory.Link(name, filepath.Base(path)); err != nil {
		if os.IsExist(err) {
			return ErrDestinationExists
		}
		return err
	}
	return nil
}

type syncWriter interface {
	io.Writer
	Sync() error
}

func stage(ctx context.Context, writer syncWriter, data []byte) error {
	for len(data) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		chunk := data[:min(len(data), 64<<10)]
		n, err := writer.Write(chunk)
		if err != nil {
			return err
		}
		if n != len(chunk) {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return writer.Sync()
}
