package tests

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Carriage represents the \r character in byte format.
const Carriage = 13

// FilterCarriage returns the input slice of bytes with \r character.
func FilterCarriage(bytes []byte) []byte {
	result := make([]byte, 0, len(bytes))
	for _, b := range bytes {
		if b != Carriage {
			result = append(result, b)
		}
	}
	return result
}

// EqualFiles compares expected and actual files.
//
// It will fail with t if one of the file cannot be read or if their content is not identical.
func EqualFiles(expected, actual string) error {
	expectedBytes, err := os.ReadFile(expected)
	if err != nil {
		return fmt.Errorf("read expected file: %w", err)
	}

	actualBytes, err := os.ReadFile(actual)
	if err != nil {
		return fmt.Errorf("read actual file: %w", err)
	}

	diffs := Diff(expected, FilterCarriage(expectedBytes), actual, FilterCarriage(actualBytes))
	if len(diffs) > 0 {
		return fmt.Errorf("there're some differences between actual and expected: %s", string(diffs))
	}
	return nil
}

// EqualDirs compares expected an actual directories (and their subdirectories).
//
// It will fail with t in case a file is missing in actual,
// a file is present in actual but not in expected
// or if the content of any file in actual is not the same as its peer in expected.
func EqualDirs(expected, actual string) error {
	expectabs, err := filepath.Abs(expected)
	if err != nil {
		return fmt.Errorf("absolute path: %w", err)
	}

	// read all files in expected directory
	expectfiles, err := readDir(expectabs, expectabs)
	if err != nil {
		return fmt.Errorf("read expected dir: %w", err)
	}

	actualabs, err := filepath.Abs(actual)
	if err != nil {
		return fmt.Errorf("absolute path: %w", err)
	}

	// read all files in actual directory
	actualfiles, err := readDir(actualabs, actualabs)
	if err != nil {
		return fmt.Errorf("read actual dir: %w", err)
	}

	// check all expected contents against actual contents
	var errs []error
	for relpath, expectedBytes := range expectfiles {
		actualBytes, ok := actualfiles[relpath]
		if !ok {
			errs = append(errs, fmt.Errorf("missing file %s from actual", relpath))
			continue
		}

		diffs := Diff(relpath, expectedBytes, relpath, actualBytes)
		if len(diffs) > 0 {
			errs = append(errs, fmt.Errorf("there're some differences between actual and expected: %s", string(diffs)))
		}
	}

	// check that there're no actual files that aren't present in expected files
	for relpath := range actualfiles {
		if _, ok := expectfiles[relpath]; !ok {
			errs = append(errs, fmt.Errorf("missing file %s from expected", relpath))
		}
	}
	return errors.Join(errs...)
}

// readDir reads a given input directory (and its subdirectories) and returns a map with filenames as keys and content (string) as values.
//
// Collision will occur in case a two files with the same name exists (between root and subdirectory).
func readDir(initialdir string, currentdir string) (map[string][]byte, error) {
	files := map[string][]byte{}

	entries, err := os.ReadDir(currentdir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	errs := make([]error, 0, len(entries))
	for _, entry := range entries {
		src := filepath.Join(currentdir, entry.Name())

		// handle directories
		if entry.IsDir() {
			sub, err := readDir(initialdir, src)
			if err != nil {
				errs = append(errs, err) // only case of error is if reading an entry fails
				continue
			}

			for filename, content := range sub {
				files[filename] = content
			}
			continue
		}

		// handle files
		bytes, err := os.ReadFile(src)
		if err != nil {
			errs = append(errs, fmt.Errorf("read file: %w", err))
			continue
		}

		abspath, err := filepath.Abs(src)
		if err != nil {
			errs = append(errs, fmt.Errorf("absolute path: %w", err))
			continue
		}

		relpath, err := filepath.Rel(initialdir, abspath)
		if err != nil {
			errs = append(errs, fmt.Errorf("relative path: %w", err))
			continue
		}
		files[relpath] = FilterCarriage(bytes)
	}
	return files, errors.Join(errs...)
}
