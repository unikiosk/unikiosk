package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/elazarl/goproxy"
	"github.com/unikiosk/unikiosk/pkg/util/file"
)

func (p *proxy) setCertificate() (err error) {
	crtE, err := file.Exist(p.config.ProxyHTTPSCertLocation)
	if err != nil || !crtE {
		return fmt.Errorf("certificate %s does not exist", p.config.ProxyHTTPSCertLocation)
	}
	keyE, err := file.Exist(p.config.ProxyHTTPSCertKeyLocation)
	if err != nil || !keyE {
		return fmt.Errorf("certificate  key %s does not exist", p.config.ProxyHTTPSCertKeyLocation)
	}

	cert, err := os.ReadFile(p.config.ProxyHTTPSCertLocation)
	if err != nil {
		return err
	}
	key, err := os.ReadFile(p.config.ProxyHTTPSCertKeyLocation)
	if err != nil {
		return err
	}

	setCA(cert, key)

	return
}

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}

	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
