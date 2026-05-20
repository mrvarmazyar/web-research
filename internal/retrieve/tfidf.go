package retrieve

import (
	"math"
	"strings"
	"unicode"
)

func tokenize(text string) []string {
	text = strings.ToLower(text)
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func tf(tokens []string) map[string]float64 {
	counts := make(map[string]int, len(tokens))
	for _, t := range tokens {
		counts[t]++
	}
	freq := make(map[string]float64, len(counts))
	n := float64(len(tokens))
	if n == 0 {
		return freq
	}
	for term, count := range counts {
		freq[term] = float64(count) / n
	}
	return freq
}

func idf(chunks [][]string) map[string]float64 {
	n := float64(len(chunks))
	docFreq := make(map[string]int)
	for _, tokens := range chunks {
		seen := make(map[string]bool)
		for _, t := range tokens {
			if !seen[t] {
				docFreq[t]++
				seen[t] = true
			}
		}
	}
	result := make(map[string]float64, len(docFreq))
	for term, df := range docFreq {
		result[term] = math.Log(n/float64(df)) + 1
	}
	return result
}

func tfidfVector(tokens []string, idfMap map[string]float64) map[string]float64 {
	tfMap := tf(tokens)
	vec := make(map[string]float64, len(tfMap))
	for term, tfVal := range tfMap {
		vec[term] = tfVal * idfMap[term]
	}
	return vec
}

func cosine(a, b map[string]float64) float64 {
	var dot, normA, normB float64
	for term, va := range a {
		dot += va * b[term]
		normA += va * va
	}
	for _, vb := range b {
		normB += vb * vb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func scoreChunks(chunks []string, query string) []float64 {
	scores := make([]float64, len(chunks))
	if len(chunks) == 0 || strings.TrimSpace(query) == "" {
		return scores
	}

	tokenized := make([][]string, len(chunks))
	for i, c := range chunks {
		tokenized[i] = tokenize(c)
	}

	queryTokens := tokenize(query)
	allDocs := append(tokenized, queryTokens)
	idfMap := idf(allDocs)

	queryVec := tfidfVector(queryTokens, idfMap)
	for i, tokens := range tokenized {
		chunkVec := tfidfVector(tokens, idfMap)
		scores[i] = cosine(queryVec, chunkVec)
	}
	return scores
}
