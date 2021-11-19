package proxy

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
	"go.uber.org/zap"
)

func (p *proxy) runHTTP(ctx context.Context) error {
	p.log.Debug("Proxy server starting up", zap.String("http", p.config.ProxyHTTPServerAddr))

	p.proxyHTTP.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Host == "" {
			fmt.Fprintln(w, "Cannot handle requests without Host header, e.g., HTTP 1.0")
			return
		}
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
		p.proxyHTTP.ServeHTTP(w, req)
	})

	p.proxyHTTP.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			p.log.Info("set headers")
			for k, v := range p.config.ProxyHeaders {
				r.Header.Set(k, v)
			}

			return r, nil
		})

	p.proxyHTTP.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*:80$"))).
		HijackConnect(func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
			defer func() {
				recover.Panic(p.log)
				client.Close()
			}()

			clientBuf := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
			remote, err := connectDial(ctx.Req.Context(), p.proxyHTTP, "tcp", req.URL.Host)
			orPanic(err)
			remoteBuf := bufio.NewReadWriter(bufio.NewReader(remote), bufio.NewWriter(remote))
			for {
				req, err := http.ReadRequest(clientBuf.Reader)
				orPanic(err)
				orPanic(req.Write(remoteBuf))
				orPanic(remoteBuf.Flush())
				resp, err := http.ReadResponse(remoteBuf.Reader, req)
				orPanic(err)
				orPanic(resp.Write(clientBuf.Writer))
				orPanic(clientBuf.Flush())
			}
		})

	// restart if panics or errors
	go func() {
		for {
			err := http.ListenAndServe(p.config.ProxyHTTPServerAddr, p.proxyHTTP)
			if err != nil {
				p.log.Debug("http proxy failed. Restarting", zap.Error(err))
				time.Sleep(time.Second)
			}
			// once failed release lock sp we can hot swap the server
			p.log.Debug("proxy restarting", zap.Error(err))
		}
	}()
	<-ctx.Done()
	return nil
}

// TODO: Drop this
func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// copied/converted from https.go
func connectDial(ctx context.Context, proxy *goproxy.ProxyHttpServer, network, addr string) (c net.Conn, err error) {
	if proxy.ConnectDial == nil {
		if proxy.Tr.DialContext != nil {
			return proxy.Tr.DialContext(ctx, network, addr)
		}
		return net.Dial(network, addr)
	}
	return proxy.ConnectDial(network, addr)
}
