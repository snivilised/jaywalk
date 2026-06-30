package bedrock

import (
	"fmt"
	"path/filepath"
	"runtime"

	nef "github.com/snivilised/nefilim"
)

// atomicWriteFile writes data to path using a safe write strategy
// that works correctly on both POSIX and Windows:
//
//  1. Write data to a temporary file in the same directory as path,
//     ensuring the rename (step 3) stays on the same filesystem and
//     avoids a cross-device move.
//  2. On POSIX: Rename(tmp, path) is atomic - the target is replaced
//     in a single kernel operation; no reader ever sees a partial file.
//     On Windows: Rename fails when the target already exists.
//     The helper removes the target first, then renames. This is a
//     best-effort two-step rather than a true atomic operation, and
//     matches the strategy used by Go's own toolchain on Windows.
//  3. Clean up the temporary file on any error path.
//
// fS is the nef.UniversalFS to write into. When nil, the real
// OS filesystem is used. Passing a luna.NewMemFS() in tests
// eliminates all real filesystem access.
func atomicWriteFile(fS nef.UniversalFS, path string, data []byte) error {
	if fS == nil {
		fS = nef.NewUniversalABS()
	}

	dir := filepath.Dir(path)
	tmpName := filepath.Join(dir, ".tmp-write-"+filepath.Base(path))

	if err := fS.WriteFile(tmpName, data, 0o600); err != nil {
		return fmt.Errorf("writing temp file %s: %w", tmpName, err)
	}

	success := false
	defer func() {
		if !success {
			_ = fS.Remove(tmpName)
		}
	}()

	if runtime.GOOS == "windows" {
		_ = fS.Remove(path)
	}

	if err := fS.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming %s to %s: %w", tmpName, path, err)
	}

	success = true
	return nil
}
