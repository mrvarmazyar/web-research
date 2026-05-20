package retrieve

import "strings"

func splitChunks(content string) []string {
	content = strings.ReplaceAll(content, "\n#", "\n\n#")
	parts := strings.Split(content, "\n\n")
	chunks := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			chunks = append(chunks, t)
		}
	}
	return chunks
}
