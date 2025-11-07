package common

import (
	"net/url"
	"strings"
)

func UrlDecode(text string) string {
	if strings.Contains(text, "%") {
		decodedStr, err := url.QueryUnescape(text)
		if err != nil {
			return text
		} else {
			return decodedStr
		}
	}

	return text
}

func UrlEncode(text string) string {
	encodedStr := url.QueryEscape(text)

	return encodedStr
}
