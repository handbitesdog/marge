package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyStatic recursively copies srcDir into destDir, preserving file mode.
// It is a no-op if srcDir does not exist.
func CopyStatic(srcDir, destDir string) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}

	err := filepath.WalkDir(srcDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, p)
		if err != nil {
			return err
		}
		dest := filepath.Join(destDir, rel)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dest, info.Mode()); err != nil {
				return err
			}
			return os.Chmod(dest, info.Mode())
		}

		return copyFile(p, dest)
	})
	if err != nil {
		return fmt.Errorf("copy static %q to %q: %w", srcDir, destDir, err)
	}

	return nil
}

func copyFile(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %q: %w", src, err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %q: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create %q: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %q to %q: %w", src, dest, err)
	}

	return os.Chmod(dest, info.Mode())
}
