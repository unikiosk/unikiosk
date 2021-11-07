package web

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/web/spaserver"
)

var _ Interface = &Service{}

type Interface interface {
	Run(ctx context.Context) error
}

type Service struct {
	log    *zap.Logger
	server *http.Server
	router *mux.Router
	config *config.Config
}

func New(
	log *zap.Logger,
	config *config.Config,
) (*Service, error) {

	s := &Service{
		log:    log,
		config: config,
	}

	s.router = s.setupRouter()

	// when running in dev mode set
	// WEB_SERVER_URI=http://localhost:3000
	if strings.HasPrefix(s.config.WebServerDir, "http://") {
		s.router.PathPrefix("/").Handler(spaserver.NewSPAReverseProxyServer(s.log, s.config.WebServerDir))
	} else {
		s.log.Info("serving static UI files", zap.String("dir", s.config.WebServerDir))

		fileSystem := http.Dir(s.config.WebServerDir)

		s.router.PathPrefix("/").Handler(spaserver.NewSPAFileServer(s.log, fileSystem))
	}

	s.server = &http.Server{
		Addr: config.WebServerURI,
		Handler: handlers.CORS(
			handlers.AllowCredentials(),
			handlers.AllowedHeaders([]string{"Content-Type"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
		)(s.router),
	}

	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	s.log.Info("Starting API Service")

	defer s.server.Shutdown(ctx)

	s.log.Info("Server will now listen", zap.String("url", s.config.WebServerURI))
	return s.server.ListenAndServe()
}

func (s *Service) setupRouter() *mux.Router {
	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	})

	return r
}
