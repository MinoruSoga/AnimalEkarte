package pet

import "github.com/animal-ekarte/backend/internal/textsearch"

func normalizeNameKana(value string) string {
	return textsearch.NormalizeKana(value)
}
