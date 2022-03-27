package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/eventer"
	"github.com/unikiosk/unikiosk/pkg/store"
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
	events eventer.Eventer
	store  store.Store
	config *config.Config
}

func New(
	log *zap.Logger,
	config *config.Config,
	events eventer.Eventer,
	store store.Store,
) (*Service, error) {

	s := &Service{
		log:    log,
		events: events,
		store:  store,
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
		Addr: config.WebServerAddr,
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

	s.log.Info("Server will now listen", zap.String("url", s.config.WebServerAddr))
	return s.server.ListenAndServe()
}

func (s *Service) setupRouter() *mux.Router {
	r := mux.NewRouter()

	code := http.HandlerFunc(s.handel)
	r.Handle("/api", code).Methods(http.MethodPost)
	r.Handle("/api", code).Methods(http.MethodGet)

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	})

	return r
}

func (s *Service) handel(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		// update
		payload := api.KioskRequest{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			_, err := fmt.Fprintf(w, "failed to parse request: %s", err.Error())
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		err = s.update(payload)
		if err != nil {
			_, err := fmt.Fprintf(w, "failed to update screen: %s", err.Error())
			if err != nil {
				fmt.Println(err.Error())
			}
		}

	case http.MethodGet:
		// get current state
		state, err := s.get()
		if err != nil {
			_, err := fmt.Fprintf(w, "failed to get screen state: %s", err.Error())
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		data, err := json.Marshal(state)
		if err != nil {
			fmt.Printf("failed to marshal: %s", err.Error())
		}

		// TODO: all return should be standart error json
		fmt.Fprint(w, data)

	default:
		fmt.Fprintf(w, "not supported")
	}

}

func (s *Service) update(payload api.KioskRequest) error {
	event := api.Event{
		Request: payload,
	}

	_, err := s.events.Emit(&eventer.EventWrapper{
		Payload: event,
	})
	return err
}

func (s *Service) get() (api.KioskResponse, error) {
	return api.KioskResponse{}, nil
}
