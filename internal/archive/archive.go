package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Entry holds one file from the SAZ archive.
type Entry struct {
	Name string
	Data []byte
}

// SessionFiles groups the main Fiddler files by session id.
type SessionFiles struct {
	SessionID int
	Request   *Entry
	Response  *Entry
	Meta      *Entry
}

// ReadSessions reads raw/*.txt and raw/*.xml files from a SAZ archive.
func ReadSessions(sazPath string) ([]SessionFiles, error) {
	zr, err := zip.OpenReader(sazPath)
	if err != nil {
		return nil, fmt.Errorf("open saz: %w", err)
	}
	defer zr.Close()

	byID := map[int]*SessionFiles{}

	for _, f := range zr.File {
		base := path.Base(f.Name)
		if !strings.HasPrefix(f.Name, "raw/") {
			continue
		}
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 {
			continue
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		entry, err := readEntry(f)
		if err != nil {
			return nil, err
		}
		sf := byID[id]
		if sf == nil {
			sf = &SessionFiles{SessionID: id}
			byID[id] = sf
		}
		switch parts[1] {
		case "c.txt":
			sf.Request = entry
		case "s.txt":
			sf.Response = entry
		case "m.xml":
			sf.Meta = entry
		}
	}

	ids := make([]int, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	out := make([]SessionFiles, 0, len(ids))
	for _, id := range ids {
		out = append(out, *byID[id])
	}
	return out, nil
}

func readEntry(f *zip.File) (*Entry, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open zip entry %s: %w", f.Name, err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read zip entry %s: %w", f.Name, err)
	}
	return &Entry{Name: f.Name, Data: data}, nil
}
