package repoquality

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/blisspixel/fartapp/internal/strictjson"
)

const (
	maxPolicyFileBytes = 4 << 20
	maxCoverageBytes   = 32 << 20
	maxMediaBytes      = 32 << 20
)

func openRegularFile(path string, maximum int64) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maximum {
		return nil, fmt.Errorf("expected a regular file of at most %d bytes: %s", maximum, path)
	}
	return os.Open(path)
}

func readFileLimited(path string, maximum int64) ([]byte, error) {
	file, err := openRegularFile(path, maximum)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("file exceeds %d bytes: %s", maximum, path)
	}
	return data, nil
}

func readPolicyJSON(path string, destination any) error {
	data, err := readFileLimited(path, maxPolicyFileBytes)
	if err != nil {
		return err
	}
	if issue := strictjson.Inspect(data, strictjson.Limits{MaximumDepth: 64, MaximumMemberNameBytes: 4096}); issue != nil {
		return fmt.Errorf("invalid policy JSON in %s: %s at %s", path, issue.Kind, issue.Path)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("invalid policy JSON in %s: %w", path, err)
	}
	return nil
}
