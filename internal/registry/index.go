// internal/registry/index.go
package registry

import (
	"encoding/json"
	"net/http"
)

type IndexEntry struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
	URL      string   `json:"url"`
}

func FetchIndex(url string) (map[string]IndexEntry, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var idx []IndexEntry
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, err
	}
	// map by name for easy lookup
	m := make(map[string]IndexEntry, len(idx))
	for _, e := range idx {
		m[e.Name] = e
	}
	return m, nil
}
