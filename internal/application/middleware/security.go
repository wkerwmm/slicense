package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	// Rate limiting
	RateLimitEnabled     bool
	RequestsPerMinute    int
	BurstSize           int
	
	// Security headers
	EnableSecurityHeaders bool
	ContentSecurityPolicy string
	HSTSMaxAge          int
	
	// CORS
	CORSEnabled         bool
	AllowedOrigins      []string
	AllowedMethods      []string
	AllowedHeaders      []string
	AllowCredentials    bool
	
	// Request validation
	MaxRequestSize      int64
	EnableRequestID     bool
	EnableRealIP        bool
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RateLimitEnabled:     true,
		RequestsPerMinute:    100,
		BurstSize:           20,
		EnableSecurityHeaders: true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		HSTSMaxAge:          31536000, // 1 year
		CORSEnabled:         true,
		AllowedOrigins:      []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:      []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:      []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials:    false,
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		EnableRequestID:     true,
		EnableRealIP:        true,
	}
}

// SecurityMiddleware provides comprehensive security features
type SecurityMiddleware struct {
	config      *SecurityConfig
	rateLimiters map[string]*rate.Limiter
}

// NewSecurityMiddleware creates a new security middleware instance
func NewSecurityMiddleware(config *SecurityConfig) *SecurityMiddleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	
	return &SecurityMiddleware{
		config:       config,
		rateLimiters: make(map[string]*rate.Limiter),
	}
}

// SecurityHeaders adds security-related HTTP headers
func (sm *SecurityMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !sm.config.EnableSecurityHeaders {
			next.ServeHTTP(w, r)
			return
		}

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy
		if sm.config.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", sm.config.ContentSecurityPolicy)
		}
		
		// HSTS (only for HTTPS)
		if r.TLS != nil && sm.config.HSTSMaxAge > 0 {
			w.Header().Set("Strict-Transport-Security", 
				fmt.Sprintf("max-age=%d; includeSubDomains", sm.config.HSTSMaxAge))
		}
		
		// Permissions Policy
		w.Header().Set("Permissions-Policy", 
			"geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), speaker=()")
		
		// Cross-Origin Embedder Policy
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		
		// Cross-Origin Opener Policy
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		
		// Cross-Origin Resource Policy
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		next.ServeHTTP(w, r)
	})
}

// RateLimit provides rate limiting functionality
func (sm *SecurityMiddleware) RateLimit(next http.Handler) http.Handler {
	if !sm.config.RateLimitEnabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := sm.getClientIP(r)
		
		// Get or create rate limiter for this IP
		limiter := sm.getRateLimiter(clientIP)
		
		// Check if request is allowed
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimit limits the size of incoming requests
func (sm *SecurityMiddleware) RequestSizeLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > sm.config.MaxRequestSize {
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}
		
		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, sm.config.MaxRequestSize)
		
		next.ServeHTTP(w, r)
	})
}

// RequestID adds a unique request ID to each request
func (sm *SecurityMiddleware) RequestID(next http.Handler) http.Handler {
	if !sm.config.EnableRequestID {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate unique request ID
		requestID := sm.generateRequestID()
		
		// Add to response headers
		w.Header().Set("X-Request-ID", requestID)
		
		// Add to request context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// RealIP extracts the real client IP from headers
func (sm *SecurityMiddleware) RealIP(next http.Handler) http.Handler {
	if !sm.config.EnableRealIP {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract real IP from headers
		realIP := sm.getClientIP(r)
		
		// Add to request context
		ctx := context.WithValue(r.Context(), "real_ip", realIP)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// CSRFProtection provides CSRF protection
func (sm *SecurityMiddleware) CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Check CSRF token
		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}
		
		// Validate CSRF token (implement your validation logic)
		if !sm.validateCSRFToken(csrfToken, r) {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// InputValidation validates and sanitizes input
func (sm *SecurityMiddleware) InputValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate URL parameters
		if !sm.validateURLParams(r) {
			http.Error(w, "Invalid URL parameters", http.StatusBadRequest)
			return
		}
		
		// Validate headers
		if !sm.validateHeaders(r) {
			http.Error(w, "Invalid headers", http.StatusBadRequest)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// SecurityLogging logs security-related events
func (sm *SecurityMiddleware) SecurityLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		next.ServeHTTP(ww, r)
		
		// Log security events
		sm.logSecurityEvent(r, ww.Status(), time.Since(start))
	})
}

// Helper methods

func (sm *SecurityMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func (sm *SecurityMiddleware) getRateLimiter(clientIP string) *rate.Limiter {
	if limiter, exists := sm.rateLimiters[clientIP]; exists {
		return limiter
	}
	
	// Create new rate limiter
	limiter := rate.NewLimiter(
		rate.Limit(sm.config.RequestsPerMinute/60.0), // requests per second
		sm.config.BurstSize,
	)
	
	sm.rateLimiters[clientIP] = limiter
	return limiter
}

func (sm *SecurityMiddleware) generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (sm *SecurityMiddleware) validateCSRFToken(token string, r *http.Request) bool {
	// Implement your CSRF token validation logic
	// This is a simplified example
	return len(token) > 0
}

func (sm *SecurityMiddleware) validateURLParams(r *http.Request) bool {
	// Validate URL parameters for potential attacks
	params := chi.URLParam(r, "*")
	
	// Check for SQL injection patterns
	sqlPatterns := []string{"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_"}
	for _, pattern := range sqlPatterns {
		if strings.Contains(strings.ToLower(params), pattern) {
			return false
		}
	}
	
	// Check for XSS patterns
	xssPatterns := []string{"<script", "javascript:", "onload=", "onerror="}
	for _, pattern := range xssPatterns {
		if strings.Contains(strings.ToLower(params), pattern) {
			return false
		}
	}
	
	return true
}

func (sm *SecurityMiddleware) validateHeaders(r *http.Request) bool {
	// Validate headers for potential attacks
	userAgent := r.Header.Get("User-Agent")
	
	// Check for suspicious user agents
	suspiciousPatterns := []string{"sqlmap", "nikto", "nmap", "masscan"}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(userAgent), pattern) {
			return false
		}
	}
	
	return true
}

func (sm *SecurityMiddleware) logSecurityEvent(r *http.Request, statusCode int, duration time.Duration) {
	// Log security events based on status code and other factors
	if statusCode >= 400 {
		// Log error responses
		sm.logSecurityEventDetails(r, statusCode, duration, "error_response")
	}
	
	// Log suspicious patterns
	if sm.isSuspiciousRequest(r) {
		sm.logSecurityEventDetails(r, statusCode, duration, "suspicious_request")
	}
}

func (sm *SecurityMiddleware) isSuspiciousRequest(r *http.Request) bool {
	// Check for suspicious patterns
	userAgent := r.Header.Get("User-Agent")
	referer := r.Header.Get("Referer")
	
	// Empty or suspicious user agent
	if userAgent == "" || len(userAgent) < 10 {
		return true
	}
	
	// Suspicious referer
	if referer != "" && !sm.isValidReferer(referer) {
		return true
	}
	
	return false
}

func (sm *SecurityMiddleware) isValidReferer(referer string) bool {
	// Validate referer against allowed origins
	for _, origin := range sm.config.AllowedOrigins {
		if strings.HasPrefix(referer, origin) {
			return true
		}
	}
	return false
}

func (sm *SecurityMiddleware) logSecurityEventDetails(r *http.Request, statusCode int, duration time.Duration, eventType string) {
	// This would integrate with your logging system
	// For now, we'll just print to console
	fmt.Printf("Security Event: %s - %s %s - Status: %d - Duration: %v - IP: %s - User-Agent: %s\n",
		eventType, r.Method, r.URL.Path, statusCode, duration, sm.getClientIP(r), r.Header.Get("User-Agent"))
}

// SetupSecurityMiddleware sets up all security middleware
func SetupSecurityMiddleware(config *SecurityConfig) []func(http.Handler) http.Handler {
	sm := NewSecurityMiddleware(config)
	
	return []func(http.Handler) http.Handler{
		sm.RequestID,
		sm.RealIP,
		sm.SecurityHeaders,
		sm.RateLimit,
		sm.RequestSizeLimit,
		sm.InputValidation,
		sm.SecurityLogging,
	}
}