package instassist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type optionEntry struct {
	Value               string `json:"value"`
	Description         string `json:"description"`
	RecommendationOrder int    `json:"recommendation_order"`
}

type optionResponse struct {
	Options []optionEntry `json:"options"`
}

func buildPrompt(userPrompt string) string {
	base := "Give me one or more concise options with short descriptions for the following: "
	schema := `Respond ONLY with JSON shaped like {"options":[{"value":"...","description":"...","recommendation_order":1}]}. No extra text.`
	return base + userPrompt + "\n" + schema
}

func parseOptions(raw string) ([]optionEntry, error) {
	var lastOpts []optionEntry
	search := raw
	for {
		idx := strings.Index(search, `{"options"`)
		if idx < 0 {
			break
		}
		segment := search[idx:]
		var resp optionResponse
		decoder := json.NewDecoder(strings.NewReader(segment))
		if err := decoder.Decode(&resp); err == nil && len(resp.Options) > 0 {
			opts := resp.Options
			sort.SliceStable(opts, func(i, j int) bool {
				oi := opts[i].RecommendationOrder
				oj := opts[j].RecommendationOrder
				if oi > 0 && oj > 0 && oi != oj {
					return oi < oj
				}
				if oi > 0 && oj <= 0 {
					return true
				}
				if oi <= 0 && oj > 0 {
					return false
				}
				return i < j
			})
			lastOpts = opts
		}
		search = search[idx+len(`{"options`):]
	}
	if len(lastOpts) > 0 {
		return lastOpts, nil
	}
	return nil, fmt.Errorf("failed to parse options JSON")
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
}

func schemaSources() (string, string, error) {
	tryPaths := []string{}

	if exe, err := os.Executable(); err == nil {
		tryPaths = append(tryPaths, filepath.Join(filepath.Dir(exe), "options.schema.json"))
	}
	if cwd, err := os.Getwd(); err == nil {
		tryPaths = append(tryPaths, filepath.Join(cwd, "options.schema.json"))
	}
	tryPaths = append(tryPaths, "/usr/local/share/insta-assist/options.schema.json")

	for _, p := range tryPaths {
		if data, err := os.ReadFile(p); err == nil {
			return p, string(data), nil
		}
	}

	// Fallback to embedded schema if available by writing to a temp file
	if len(embeddedSchema) > 0 {
		tmp, err := os.CreateTemp("", "insta-options-schema-*.json")
		if err != nil {
			return "", "", fmt.Errorf("failed to create temp schema file: %w", err)
		}
		if _, err := tmp.Write(embeddedSchema); err != nil {
			tmp.Close()
			return "", "", fmt.Errorf("failed to write temp schema file: %w", err)
		}
		if err := tmp.Close(); err != nil {
			return "", "", fmt.Errorf("failed to close temp schema file: %w", err)
		}
		return tmp.Name(), string(embeddedSchema), nil
	}

	return "", "", fmt.Errorf("options.schema.json not found in executable directory, working directory, or /usr/local/share/insta-assist")
}
