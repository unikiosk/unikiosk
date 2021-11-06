package spaserver

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type SPAFileServer struct {
	log        *zap.Logger
	fileSystem http.FileSystem
	fileServer http.Handler
}

func NewSPAFileServer(log *zap.Logger, fileSystem http.FileSystem) *SPAFileServer {
	return &SPAFileServer{
		log:        log,
		fileSystem: fileSystem,
		fileServer: http.FileServer(fileSystem),
	}
}

func (s *SPAFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		s.log.Error("serve http error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if f, err := s.fileSystem.Open(path); err == nil {
		if err = f.Close(); err != nil {
			s.log.Error("serve http error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.fileServer.ServeHTTP(w, r)
	} else if os.IsNotExist(err) {
		r.URL.Path = ""
		s.fileServer.ServeHTTP(w, r)
		return
	} else {
		s.log.Error("file system open", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// SPAReverseProxyServer is used for local development or in theory it could be used
// if the frontend is running on a different container
type SPAReverseProxyServer struct {
	reverseProxy *httputil.ReverseProxy
}

func NewSPAReverseProxyServer(log *zap.Logger, frontend string) *SPAReverseProxyServer {
	u, err := url.Parse(frontend)
	if err != nil {
		log.Fatal("failed to parse frontend URL", zap.String("url", frontend))
	}

	return &SPAReverseProxyServer{
		reverseProxy: httputil.NewSingleHostReverseProxy(u),
	}
}

func (s *SPAReverseProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.reverseProxy.ServeHTTP(w, r)
}
