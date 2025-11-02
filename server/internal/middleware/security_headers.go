package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds comprehensive security headers to all responses
// This middleware implements OWASP security best practices
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Content Security Policy (CSP)
			// Helps prevent XSS attacks by controlling which resources can be loaded
			c.Response().Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
					"style-src 'self' 'unsafe-inline'; "+
					"img-src 'self' data: https:; "+
					"font-src 'self' data:; "+
					"connect-src 'self'; "+
					"frame-ancestors 'none'; "+
					"base-uri 'self'; "+
					"form-action 'self'")

			// HTTP Strict Transport Security (HSTS)
			// Forces browsers to only connect via HTTPS
			// max-age=31536000 (1 year), includeSubDomains, preload
			c.Response().Header().Set("Strict-Transport-Security",
				"max-age=31536000; includeSubDomains; preload")

			// X-Frame-Options
			// Prevents clickjacking attacks by preventing page from being embedded in iframe
			c.Response().Header().Set("X-Frame-Options", "DENY")

			// X-Content-Type-Options
			// Prevents MIME-sniffing attacks
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// X-XSS-Protection
			// Enables browser's built-in XSS protection
			// Note: This header is deprecated in modern browsers but included for legacy support
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer-Policy
			// Controls how much referrer information is sent with requests
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy (formerly Feature-Policy)
			// Controls which browser features can be used
			c.Response().Header().Set("Permissions-Policy",
				"geolocation=(), microphone=(), camera=(), payment=()")

			// X-Permitted-Cross-Domain-Policies
			// Prevents Adobe Flash and PDF from loading cross-domain content
			c.Response().Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			// Cache-Control for sensitive data
			// Note: This is a general setting; specific routes may override
			c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Response().Header().Set("Pragma", "no-cache")

			return next(c)
		}
	}
}

// PublicCacheHeaders sets appropriate cache headers for public, cacheable resources
// Use this for static assets that can be safely cached
func PublicCacheHeaders(maxAge int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
			return next(c)
		}
	}
}
