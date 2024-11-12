package common

import (
	"fmt"
	"net/url"
)

func UrlDecode(text string) string {
	decodedStr, err := url.QueryUnescape(text)
	if err != nil {
		fmt.Println("Error decoding URL:", err)
		return err.Error()
	} else {
		return decodedStr
	}
}

func UrlEncode(text string) string {
	encodedStr := url.QueryEscape(text)

	return encodedStr
}
