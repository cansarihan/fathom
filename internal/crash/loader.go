package crash

import (
	"os"
	"path/filepath"
)

// Input is one recorded failing input on disk: the target it belongs to (taken
// from its directory), its path, and the decoded corpus record.
type Input struct {
	Target string `json:"target"`
	Path   string `json:"path"`
	Record Record `json:"record"`
}

// LoadInputs walks a target's testdata/fuzz tree and decodes every recorded
// failing input it finds. Files that are not corpus files are skipped.
func LoadInputs(root string) ([]Input, error) {
	var inputs []Input
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path) // #nosec G304 -- walking the target's own testdata
		if readErr != nil {
			return nil
		}
		rec, decodeErr := DecodeRecord(data)
		if decodeErr != nil {
			return nil
		}
		inputs = append(inputs, Input{
			Target: filepath.Base(filepath.Dir(path)),
			Path:   path,
			Record: rec,
		})
		return nil
	})
	return inputs, err
}
