package api

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/cors"
	"github.com/sebest/xff"
	"github.com/sirupsen/logrus"
)

const (
	audHeaderName  = "X-JWT-AUD"
	defaultVersion = "unknown version"
)

// Config is the main API config
type Config struct {
	Host             string `default:"localhost"`
	Port             int    `default:"8085"`
	JwtSecret        string `required:"true" split_words:"true"`
	MetricCollectors string `required:"false" default:"meminfo,loadavg,cpu"`
}

func (c *Config) GetEnabledCollectors() []string {
	splits := strings.Split(c.MetricCollectors, ",")
	filtered := make([]string, 0)
	for _, c := range splits {
		if len(strings.TrimSpace(c)) == 0 {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

// API is the main REST API
type API struct {
	handler http.Handler
	config  *Config
	version string
}

// ListenAndServe starts the REST API
func (a *API) ListenAndServe(hostAndPort string) {
	log := logrus.WithField("component", "api")
	server := &http.Server{
		Addr:    hostAndPort,
		Handler: a.handler,
	}

	done := make(chan struct{})
	defer close(done)
	go func() {
		waitForTermination(log, done)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.WithError(err).Fatal("http server listen failed")
	}
}

// WaitForShutdown blocks until the system signals termination or done has a value
func waitForTermination(log logrus.FieldLogger, done <-chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-signals:
		log.Infof("Triggering shutdown from signal %s", sig)
	case <-done:
		log.Infof("Shutting down...")
	}
}

// NewAPI instantiates a new REST API
func NewAPI(config *Config) *API {
	return NewAPIWithVersion(config, defaultVersion)
}

// NewAPIWithVersion creates a new REST API using the specified version
func NewAPIWithVersion(config *Config, version string) *API {
	api := &API{config: config, version: version}
	metrics, err := NewMetrics(config.GetEnabledCollectors()); if err != nil {
		panic(fmt.Sprintf("Couldn't initialize metrics: %+v", err))
	}

	xffmw, _ := xff.Default()

	r := chi.NewRouter()
	r.Use(xffmw.Handler)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// unauthenticated
	r.Group(func(r chi.Router) {
		r.Method("GET", "/metrics", metrics.GetHandler())
	})

	// private endpoints
	r.Group(func(r chi.Router) {
		r.Use(api.AuthHandler)
		r.Method("GET", "/health", ErrorHandlingWrapper(api.HealthCheck))

		r.Route("/", func(r chi.Router) {
			r.Route("/test", func(r chi.Router) {
				r.Method("GET", "/", ErrorHandlingWrapper(api.TestGet))
			})

			r.Route("/service", func(r chi.Router) {
				// applications are kong, pglisten, postgrest, goauth, realtime, adminapi, all
				r.Route("/restart", func(r chi.Router) {
					r.Method("GET", "/", ErrorHandlingWrapper(api.RestartServices))
					r.Method("GET", "/{application}", ErrorHandlingWrapper(api.RestartServices))
				})
				r.Method("GET", "/reboot", ErrorHandlingWrapper(api.RebootMachine))
			})

			// applications are kong, pglisten, postgrest, goauth, realtime, adminapi
			r.Route("/config/{application}", func(r chi.Router) {
				r.Method("GET", "/", ErrorHandlingWrapper(api.GetFileContents))
				r.Method("POST", "/", ErrorHandlingWrapper(api.SetFileContents))
			})

			// applications are kong, pglisten, postgrest, goauth, realtime
			r.Route("/logs/{application}/{type}/{n:[0-9]*}", func(r chi.Router) {
				r.Method("GET", "/", ErrorHandlingWrapper(api.GetLogContents))
			})

			r.Route("/cert", func(r chi.Router) {
				r.Method("POST", "/", ErrorHandlingWrapper(api.UpdateCert))
			})
	})
	})

	corsHandler := cors.New(cors.Options{
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", audHeaderName},
		AllowCredentials: true,
	})

	api.handler = corsHandler.Handler(chi.ServerBaseContext(context.Background(), r))
	return api
}

// HealthCheck returns basic information for status purposes
func (a *API) HealthCheck(w http.ResponseWriter, r *http.Request) error {
	return sendJSON(w, http.StatusOK, map[string]string{
		"version":     a.version,
		"name":        "supabase-admin-api",
		"description": "supabase-admin-api is an api to manage KPS",
	})
}
