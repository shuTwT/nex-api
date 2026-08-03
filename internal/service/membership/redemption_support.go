package membership

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

func generateCodes(count int) ([]string, error) {
	codes := make([]string, 0, count)
	seen := make(map[string]struct{}, count)
	for len(codes) < count {
		code, err := generateCode()
		if err != nil {
			return nil, fmt.Errorf("generate redemption code: %w", err)
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	return codes, nil
}

func generateCode() (string, error) {
	raw := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", err
	}
	code := make([]byte, len(raw))
	for index, value := range raw {
		code[index] = redemptionAlphabet[int(value)%len(redemptionAlphabet)]
	}
	return string(code), nil
}

func normalizeCode(input string) string { return strings.ToUpper(strings.TrimSpace(input)) }

func normalizeRedemptionFilter(filter RedemptionListFilter) RedemptionListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}
	return filter
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func usedLabel(used bool) string {
	if used {
		return "已使用"
	}
	return "未使用"
}
