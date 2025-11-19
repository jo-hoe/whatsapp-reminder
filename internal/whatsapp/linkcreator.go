package whatsapp

import (
	"net/url"
	"regexp"
	"strings"
)

// creates links as described here
// https://faq.whatsapp.com/iphone/how-to-link-to-whatsapp-from-a-different-app/?lang=en
// full example:
// https://wa.me/15551234567?text=I'm%20interested%20in%20your%20car%20for%20sale
// example with only phone number
// https://wa.me/15551234567
// example with only text
// https://wa.me/?text=urlencodedtext

const baseURL = "https://wa.me/"

func CreateWhatsappLink(phoneNumber string, messageText string) (link string) {
	var stringBuilder strings.Builder
	stringBuilder.WriteString(baseURL)
	phoneNumberWithoutWhitespace := removeWhiteSpaces(phoneNumber)
	stringBuilder.WriteString(url.PathEscape(phoneNumberWithoutWhitespace))
	if len(messageText) > 0 {
		stringBuilder.WriteString("?text=")
		stringBuilder.WriteString(url.PathEscape(messageText))
	}
	return stringBuilder.String()
}

func removeWhiteSpaces(str string) string {
	if str == "" {
		return ""
	}
	//A regular expression that matches one or more whitespace characters
	reg := regexp.MustCompile(`\s+`)
	return reg.ReplaceAllString(str, "")
}
