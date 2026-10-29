package mlcext

// RefNos (document ids such as FB-QM-001): the next free one, proposed
// from the documents of a folder.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var refNoRe = regexp.MustCompile(`^([A-Za-z]+)-([A-Za-z0-9]+)-(\d+)$`)

// SuggestRefNo proposes the next free RefNo for prefix ("FB") and area
// ("QM") after the ones in existing: FB-QM-002 → FB-QM-003, keeping the
// number of digits (at least three). Only a proposal: whether a RefNo is
// free among the approved documents only the Sensor Monitor knows.
func SuggestRefNo(prefix, area string, existing []string) string {
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	area = strings.ToUpper(strings.TrimSpace(area))
	if prefix == "" || area == "" {
		return ""
	}
	maxNo, width := 0, 3
	for _, e := range existing {
		m := refNoRe.FindStringSubmatch(strings.TrimSpace(e))
		if m == nil || !strings.EqualFold(m[1], prefix) || !strings.EqualFold(m[2], area) {
			continue
		}
		if n, err := strconv.Atoi(m[3]); err == nil && n > maxNo {
			maxNo, width = n, max(width, len(m[3]))
		}
	}
	return fmt.Sprintf("%s-%s-%0*d", prefix, area, width, maxNo+1)
}

// RefNosInDir collects the RefNos of the Markdown files below dir (at most
// maxFiles files, hidden directories skipped), for SuggestRefNo.
func RefNosInDir(dir string, maxFiles int) []string {
	var refs []string
	n := 0
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != dir && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		if n++; n > maxFiles {
			return filepath.SkipAll
		}
		b, err := os.ReadFile(path)
		if err == nil {
			if ref := DocumentMeta(string(b)).DocID; ref != "" {
				refs = append(refs, ref)
			}
		}
		return nil
	})
	return refs
}
