package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets defensive HTTP response headers on every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent the page from being embedded in iframes (clickjacking)
		c.Header("X-Frame-Options", "DENY")
		// Prevent MIME-type sniffing attacks
		c.Header("X-Content-Type-Options", "nosniff")
		// Legacy XSS filter for older browsers
		c.Header("X-XSS-Protection", "1; mode=block")
		// Control referrer information sent on navigations
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Disable unused browser features
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// Content Security Policy restricts resource loading origins.
		// 'unsafe-inline' for style-src is a pragmatic allowance for Tailwind.
		c.Header("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")
		// Remove server version disclosure
		c.Header("Server", "")
		c.Next()
	}
}
