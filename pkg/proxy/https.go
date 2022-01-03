package proxy

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/elazarl/goproxy"
	"go.uber.org/zap"
)

func (p *proxy) runHTTPS(ctx context.Context) error {
	p.log.Debug("Proxy server starting up", zap.String("http", p.config.ProxyHTTPSServerAddr))

	p.proxyHTTPS.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			//p.log.Info("set headers")
			for k, v := range p.config.ProxyHeaders {
				r.Header.Set(k, v)
			}

			return r, nil
		})

	p.proxyHTTPS.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)

	// restart if panics or errors
	go func() {
		for {
			err := http.ListenAndServe(p.config.ProxyHTTPSServerAddr, p.proxyHTTPS)
			if err != nil {
				p.log.Debug("https proxy failed. Restarting", zap.Error(err))
				time.Sleep(time.Second)
			}
			// once failed release lock sp we can hot swap the server
			p.log.Debug("proxy restarting", zap.Error(err))
		}
	}()
	<-ctx.Done()
	return nil
}
