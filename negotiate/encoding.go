package negotiate

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

const AcceptEncodingK = "Accept-Encoding"
const identityEncoding = "identity"

func EncodingH(h http.Header, availableEncodings []string) string {
	return Encoding(h[AcceptEncodingK], availableEncodings)
}

func Encoding(acceptEncodings []string, availableEncodings []string) string {
	if len(availableEncodings) == 0 {
		return ""
	}

	// if the Accept-Encoding header is not found, then only send back "identity".
	if len(acceptEncodings) == 0 {
		if !slices.Contains(availableEncodings, identityEncoding) {
			return ""
		}
		return identityEncoding
	}

	wildcardQ := findEncodingWeight(acceptEncodings, "*")
	identityQ := 0.0

	maxV := ""
	maxQ := 0.0

	for _, curr := range availableEncodings {
		q := findEncodingWeightWithFallback(acceptEncodings, curr, wildcardQ)
		if curr == identityEncoding {
			identityQ = q
		}

		if q == -1.0 {
			continue
		}

		// 1.0 is the maximum, so return that immediately if found
		if q == 1.0 {
			return curr
		}

		if q > maxQ {
			maxQ = q
			maxV = curr
		}
	}

	if maxQ == 0.0 {
		if identityQ != 0.0 {
			return identityEncoding
		}
		return ""
	}

	return maxV
}

// findEncodingWeight returns the q value found for v.
// if v is not found, or there was an issue parsing, -1.0 is returned.
func findEncodingWeight(acceptEncodings []string, v string) float64 {
	for _, el := range acceptEncodings {
		for e := range strings.SplitSeq(el, ",") {
			encoding, rest, found := strings.Cut(e, ";")
			encoding = strings.TrimSpace(encoding)
			if encoding != v {
				continue
			}
			return extractWeight(rest, found)
		}
	}
	return -1.0 // not found
}

func extractWeight(rest string, found bool) float64 {
	if !found {
		return 1.0 // default weight of 1.0
	}
	_, weightStr, found := strings.Cut(rest, "q=")
	if !found {
		return -1.0 // invalid format
	}
	weight, err := strconv.ParseFloat(strings.TrimSpace(weightStr), 64)
	if err != nil {
		return -1.0 // invalid float
	}
	return max(min(weight, 1.0), 0.0)
}

func findEncodingWeightWithFallback(acceptEncodings []string, v string, fallback float64) float64 {
	w := findEncodingWeight(acceptEncodings, v)
	if w == -1.0 {
		return fallback
	}
	return w
}
