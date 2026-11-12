package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

// I18nMiddleware internationalization middleware
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the language from the request header
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = "en" // Default to English
		}

		// Parse the language tag
		tag, err := language.Parse(lang)
		if err != nil {
			tag = language.English
		}

		// Set the language in the context
		c.Set("language", tag.String())

		// Set the response header
		c.Header("Content-Language", tag.String())

		c.Next()
	}
}
