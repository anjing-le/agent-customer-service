package store

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

const maxKnowledgeEvidence = 3

type knowledgeCandidate struct {
	item  KnowledgeArticle
	score float64
	index int
}

func rankKnowledge(query string, items []KnowledgeArticle, limit int) []KnowledgeArticle {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	if limit <= 0 {
		limit = maxKnowledgeEvidence
	}

	normalizedQuery := normalizeRetrievalText(query)
	queryTerms := retrievalTerms(query)
	candidates := make([]knowledgeCandidate, 0, len(items))
	for index, item := range items {
		score, reason := scoreKnowledgeArticle(normalizedQuery, queryTerms, item)
		if score < 20 {
			continue
		}
		copyItem := item
		copyItem.RetrievalScore = math.Round(score*10) / 10
		copyItem.RetrievalReason = reason
		candidates = append(candidates, knowledgeCandidate{item: copyItem, score: score, index: index})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		leftTrust := trustWeight(candidates[i].item.TrustLevel)
		rightTrust := trustWeight(candidates[j].item.TrustLevel)
		if leftTrust != rightTrust {
			return leftTrust > rightTrust
		}
		if candidates[i].item.UpdatedAt != candidates[j].item.UpdatedAt {
			return candidates[i].item.UpdatedAt > candidates[j].item.UpdatedAt
		}
		return candidates[i].index < candidates[j].index
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]KnowledgeArticle, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, candidate.item)
	}
	return result
}

func scoreKnowledgeArticle(normalizedQuery string, queryTerms map[string]struct{}, item KnowledgeArticle) (float64, string) {
	score := 0.0
	reasons := make([]string, 0, 4)
	title := normalizeRetrievalText(item.Title)
	category := normalizeRetrievalText(item.Category)

	if title != "" && strings.Contains(normalizedQuery, title) {
		score += 80
		reasons = append(reasons, "title")
	}
	if category != "" && strings.Contains(normalizedQuery, category) {
		score += 34
		reasons = append(reasons, "category")
	}
	for _, tag := range item.Tags {
		normalizedTag := normalizeRetrievalText(tag)
		if normalizedTag != "" && strings.Contains(normalizedQuery, normalizedTag) {
			score += 36
			reasons = append(reasons, "tag:"+tag)
		}
	}

	docText := strings.Join([]string{item.Title, item.Category, item.Content, strings.Join(item.Tags, " ")}, " ")
	docTerms := retrievalTerms(docText)
	overlap := 0
	for term := range queryTerms {
		if _, ok := docTerms[term]; ok {
			overlap++
		}
	}
	if len(queryTerms) > 0 && overlap > 0 {
		score += (float64(overlap) / float64(len(queryTerms))) * 35
		reasons = append(reasons, "semantic-overlap")
	}
	if score > 0 {
		score += trustWeight(item.TrustLevel)
	}
	return score, strings.Join(uniqueStrings(reasons), ", ")
}

func normalizeRetrievalText(text string) string {
	var b strings.Builder
	lastSpace := false
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteRune(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func retrievalTerms(text string) map[string]struct{} {
	terms := make(map[string]struct{})
	for _, field := range strings.Fields(normalizeRetrievalText(text)) {
		addRetrievalFieldTerms(terms, field)
	}
	return terms
}

func addRetrievalFieldTerms(terms map[string]struct{}, field string) {
	var han []rune
	flushHan := func() {
		if len(han) >= 2 {
			for i := 0; i < len(han)-1; i++ {
				terms[string(han[i:i+2])] = struct{}{}
			}
		}
		han = han[:0]
	}

	var ascii strings.Builder
	flushASCII := func() {
		value := ascii.String()
		if len(value) >= 2 {
			terms[value] = struct{}{}
		}
		ascii.Reset()
	}

	for _, r := range field {
		if unicode.Is(unicode.Han, r) {
			flushASCII()
			han = append(han, r)
			continue
		}
		flushHan()
		ascii.WriteRune(r)
	}
	flushHan()
	flushASCII()
}

func trustWeight(value string) float64 {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "HIGH":
		return 4
	case "MEDIUM":
		return 2
	default:
		return 0
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
