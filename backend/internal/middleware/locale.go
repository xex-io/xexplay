package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const ContextKeyLocale = "locale"

var supportedLocales = map[string]bool{
	"en": true,
	"fa": true,
	"ar": true,
	"tr": true,
	"es": true,
	"fr": true,
}

// Locale parses the Accept-Language header and sets the locale in context.
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := "en" // default

		// Check Accept-Language header first
		acceptLang := c.GetHeader("Accept-Language")
		if acceptLang != "" {
			// Parse simple language tag (e.g., "fa", "en-US")
			lang := strings.SplitN(acceptLang, ",", 2)[0]
			lang = strings.SplitN(strings.TrimSpace(lang), "-", 2)[0]
			lang = strings.ToLower(lang)
			if supportedLocales[lang] {
				locale = lang
			}
		}

		c.Set(ContextKeyLocale, locale)
		c.Next()
	}
}
