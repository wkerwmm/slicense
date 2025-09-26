package monitoring

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  prometheus.Gauge

	// License metrics
	LicenseVerificationsTotal *prometheus.CounterVec
	LicenseActivationsTotal   *prometheus.CounterVec
	LicenseCreationsTotal     *prometheus.CounterVec
	LicenseDeletionsTotal     *prometheus.CounterVec

	// User metrics
	UserRegistrationsTotal *prometheus.CounterVec
	UserLoginsTotal        *prometheus.CounterVec
	UserLoginsFailed       *prometheus.CounterVec

	// Database metrics
	DatabaseConnectionsActive prometheus.Gauge
	DatabaseConnectionsIdle   prometheus.Gauge
	DatabaseQueryDuration     *prometheus.HistogramVec
	DatabaseErrorsTotal       *prometheus.CounterVec

	// Cache metrics
	CacheHitsTotal   *prometheus.CounterVec
	CacheMissesTotal *prometheus.CounterVec
	CacheOperations  *prometheus.CounterVec

	// System metrics
	SystemMemoryUsage    prometheus.Gauge
	SystemCPUUsage       prometheus.Gauge
	SystemGoroutines     prometheus.Gauge
	SystemGCPause        *prometheus.HistogramVec

	// Business metrics
	ActiveLicensesTotal    prometheus.Gauge
	ExpiredLicensesTotal   prometheus.Gauge
	ActiveUsersTotal       prometheus.Gauge
	AuditLogsTotal         prometheus.Counter
}

// NewMetrics creates a new Metrics instance with all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// License metrics
		LicenseVerificationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "license_verifications_total",
				Help: "Total number of license verifications",
			},
			[]string{"product", "result"},
		),
		LicenseActivationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "license_activations_total",
				Help: "Total number of license activations",
			},
			[]string{"product", "result"},
		),
		LicenseCreationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "license_creations_total",
				Help: "Total number of license creations",
			},
			[]string{"product"},
		),
		LicenseDeletionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "license_deletions_total",
				Help: "Total number of license deletions",
			},
			[]string{"product"},
		),

		// User metrics
		UserRegistrationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "user_registrations_total",
				Help: "Total number of user registrations",
			},
			[]string{"result"},
		),
		UserLoginsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "user_logins_total",
				Help: "Total number of user logins",
			},
			[]string{"result"},
		),
		UserLoginsFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "user_logins_failed_total",
				Help: "Total number of failed user logins",
			},
			[]string{"reason"},
		),

		// Database metrics
		DatabaseConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "database_connections_active",
				Help: "Number of active database connections",
			},
		),
		DatabaseConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "database_connections_idle",
				Help: "Number of idle database connections",
			},
		),
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		DatabaseErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_errors_total",
				Help: "Total number of database errors",
			},
			[]string{"operation", "error_type"},
		),

		// Cache metrics
		CacheHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type", "key_pattern"},
		),
		CacheMissesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type", "key_pattern"},
		),
		CacheOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_operations_total",
				Help: "Total number of cache operations",
			},
			[]string{"operation", "cache_type", "result"},
		),

		// System metrics
		SystemMemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		SystemCPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),
		SystemGoroutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_goroutines_total",
				Help: "Current number of goroutines",
			},
		),
		SystemGCPause: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "system_gc_pause_seconds",
				Help:    "GC pause duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"gc_type"},
		),

		// Business metrics
		ActiveLicensesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_licenses_total",
				Help: "Total number of active licenses",
			},
		),
		ExpiredLicensesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "expired_licenses_total",
				Help: "Total number of expired licenses",
			},
		),
		ActiveUsersTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_users_total",
				Help: "Total number of active users",
			},
		),
		AuditLogsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "audit_logs_total",
				Help: "Total number of audit log entries",
			},
		),
	}
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordLicenseVerification records license verification metrics
func (m *Metrics) RecordLicenseVerification(product, result string) {
	m.LicenseVerificationsTotal.WithLabelValues(product, result).Inc()
}

// RecordLicenseActivation records license activation metrics
func (m *Metrics) RecordLicenseActivation(product, result string) {
	m.LicenseActivationsTotal.WithLabelValues(product, result).Inc()
}

// RecordLicenseCreation records license creation metrics
func (m *Metrics) RecordLicenseCreation(product string) {
	m.LicenseCreationsTotal.WithLabelValues(product).Inc()
}

// RecordLicenseDeletion records license deletion metrics
func (m *Metrics) RecordLicenseDeletion(product string) {
	m.LicenseDeletionsTotal.WithLabelValues(product).Inc()
}

// RecordUserRegistration records user registration metrics
func (m *Metrics) RecordUserRegistration(result string) {
	m.UserRegistrationsTotal.WithLabelValues(result).Inc()
}

// RecordUserLogin records user login metrics
func (m *Metrics) RecordUserLogin(result string) {
	m.UserLoginsTotal.WithLabelValues(result).Inc()
}

// RecordUserLoginFailed records failed user login metrics
func (m *Metrics) RecordUserLoginFailed(reason string) {
	m.UserLoginsFailed.WithLabelValues(reason).Inc()
}

// RecordDatabaseQuery records database query metrics
func (m *Metrics) RecordDatabaseQuery(operation, table string, duration time.Duration) {
	m.DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordDatabaseError records database error metrics
func (m *Metrics) RecordDatabaseError(operation, errorType string) {
	m.DatabaseErrorsTotal.WithLabelValues(operation, errorType).Inc()
}

// RecordCacheHit records cache hit metrics
func (m *Metrics) RecordCacheHit(cacheType, keyPattern string) {
	m.CacheHitsTotal.WithLabelValues(cacheType, keyPattern).Inc()
}

// RecordCacheMiss records cache miss metrics
func (m *Metrics) RecordCacheMiss(cacheType, keyPattern string) {
	m.CacheMissesTotal.WithLabelValues(cacheType, keyPattern).Inc()
}

// RecordCacheOperation records cache operation metrics
func (m *Metrics) RecordCacheOperation(operation, cacheType, result string) {
	m.CacheOperations.WithLabelValues(operation, cacheType, result).Inc()
}

// UpdateSystemMetrics updates system-level metrics
func (m *Metrics) UpdateSystemMetrics() {
	// Update goroutine count
	m.SystemGoroutines.Set(float64(runtime.NumGoroutine()))

	// Update memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.SystemMemoryUsage.Set(float64(memStats.Alloc))

	// Update GC pause metrics
	m.SystemGCPause.WithLabelValues("gc").Observe(float64(memStats.PauseNs[(memStats.NumGC+255)%256]) / 1e9)
}

// UpdateBusinessMetrics updates business-level metrics
func (m *Metrics) UpdateBusinessMetrics(ctx context.Context, db *sql.DB) {
	// Update active licenses count
	var activeLicenses int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM licenses WHERE expires_at IS NULL OR expires_at > NOW()").Scan(&activeLicenses)
	if err == nil {
		m.ActiveLicensesTotal.Set(float64(activeLicenses))
	}

	// Update expired licenses count
	var expiredLicenses int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM licenses WHERE expires_at IS NOT NULL AND expires_at <= NOW()").Scan(&expiredLicenses)
	if err == nil {
		m.ExpiredLicensesTotal.Set(float64(expiredLicenses))
	}

	// Update active users count
	var activeUsers int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Accounts WHERE last_login > DATE_SUB(NOW(), INTERVAL 30 DAY)").Scan(&activeUsers)
	if err == nil {
		m.ActiveUsersTotal.Set(float64(activeUsers))
	}
}

// StartMetricsUpdater starts a goroutine to periodically update metrics
func (m *Metrics) StartMetricsUpdater(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.UpdateSystemMetrics()
			m.UpdateBusinessMetrics(ctx, db)
		}
	}
}