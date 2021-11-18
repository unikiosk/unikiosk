package proxy

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/cybozu-go/transocks"
	"github.com/davecgh/go-spew/spew"
	"github.com/elazarl/goproxy"
	"github.com/inconshreveable/go-vhost"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/util/recover"
)

type Proxy interface {
	Run(ctx context.Context) error
}

type proxy struct {
	config *config.Config
	proxy  *goproxy.ProxyHttpServer

	log *zap.Logger
}

func New(ctx context.Context, log *zap.Logger, config *config.Config) (*proxy, error) {

	p := proxy{
		log:    log,
		config: config,
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	p.proxy = proxy

	return &p, nil
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

func (p *proxy) Run(ctx context.Context) error {
	p.log.Debug("Proxy server starting up", zap.String("http", p.config.ProxyHTTPServerAddr), zap.String("https", p.config.ProxyHTTPSServerAddr))

	p.proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Host == "" {
			fmt.Fprintln(w, "Cannot handle requests without Host header, e.g., HTTP 1.0")
			return
		}
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
		p.proxy.ServeHTTP(w, req)
	})

	p.proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			p.log.Info("set headers")
			for k, v := range p.config.ProxyHeaders {
				r.Header.Set(k, v)
			}

			return r, nil
		})

	p.proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)

	p.proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*:80$"))).
		HijackConnect(func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
			defer func() {
				recover.Panic(p.log)
				client.Close()
			}()

			clientBuf := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
			remote, err := connectDial(ctx.Req.Context(), p.proxy, "tcp", req.URL.Host)
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

	go func() {
		for {
			err := http.ListenAndServe(p.config.ProxyHTTPServerAddr, p.proxy)
			if err != nil {
				p.log.Debug("http proxy failed. Restarting", zap.Error(err))
				time.Sleep(time.Second)
			}
			// once failed release lock sp we can hot swap the server
			p.log.Debug("proxy restarting", zap.Error(err))
		}
	}()

	addr, err := net.ResolveTCPAddr("tcp", p.config.ProxyHTTPSServerAddr)
	if err != nil {
		panic(err)
	}

	// listen to the TLS ClientHello but make it a CONNECT request instead
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Error listening for https connections - %v", err)
	}

	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Printf("Error accepting new connection - %v", err)
			continue
		}

		go func(c *net.TCPConn) {
			tlsConn, err := vhost.TLS(c)
			if err != nil {
				log.Printf("Error accepting new connection - %v", err)
				return
			}
			originalDest := ""
			if tlsConn.Host() == "" {
				log.Printf("Cannot support non-SNI enabled clients. Fallback using `SO_ORIGINAL_DST` or `IP6T_SO_ORIGINAL_DST`")
				origAddr, err := transocks.GetOriginalDST(c)
				if err != nil {
					log.Printf("GetOriginalDST failed - %s", err.Error())
					return
				}
				originalDest = origAddr.String()
				// TODO getting domain from origAddr, then check whether we should use proxy or not
			} else {
				originalDest = tlsConn.Host()
			}
			spew.Dump(originalDest)

			connectReq := &http.Request{
				Method: "CONNECT",
				URL: &url.URL{
					Opaque: originalDest,
					Host:   net.JoinHostPort(originalDest, "443"),
				},
				Host:       originalDest,
				Header:     make(http.Header),
				RemoteAddr: c.RemoteAddr().String(),
			}
			resp := dumbResponseWriter{tlsConn}
			p.proxy.ServeHTTP(resp, connectReq)
		}(c)
	}

}

func (p *proxy) Stop(ctx context.Context) error {
	p.log.Info("stopping proxy")
	// TODO: implement
	return nil
}
