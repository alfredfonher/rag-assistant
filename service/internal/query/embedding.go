package query

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

type EmbeddingProvider interface {
	EmbedDocument(ctx context.Context, text string) ([]float64, error)
	EmbedQuery(ctx context.Context, text string) ([]float64, error)
}

type DeterministicEmbeddingProvider struct {
	dimension int
}

func NewDeterministicEmbeddingProvider(dimension int) *DeterministicEmbeddingProvider {
	if dimension <= 0 {
		dimension = 16
	}

	return &DeterministicEmbeddingProvider{dimension: dimension}
}

func (p *DeterministicEmbeddingProvider) EmbedDocument(_ context.Context, text string) ([]float64, error) {
	return p.embed(text), nil
}

func (p *DeterministicEmbeddingProvider) EmbedQuery(_ context.Context, text string) ([]float64, error) {
	return p.embed(text), nil
}

func (p *DeterministicEmbeddingProvider) embed(text string) []float64 {
	vector := make([]float64, p.dimension)
	for _, token := range normalizeEmbeddingTokens(text) {
		hash := fnv.New32a()
		_, _ = hash.Write([]byte(token))
		vector[int(hash.Sum32()%uint32(p.dimension))]++
	}

	return normalizeVector(vector)
}

func normalizeEmbeddingTokens(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func normalizeVector(vector []float64) []float64 {
	var sumSquares float64
	for _, value := range vector {
		sumSquares += value * value
	}

	if sumSquares == 0 {
		return vector
	}

	scale := 1 / math.Sqrt(sumSquares)
	for i := range vector {
		vector[i] *= scale
	}

	return vector
}
