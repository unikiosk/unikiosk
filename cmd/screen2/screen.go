package main

import (
	"github.com/webview/webview"
)

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("ProxyTest")
	w.SetSize(800, 600, webview.HintNone)
	w.Proxy("https://0.0.0.0:44933", nil)
	w.Proxy("http://0.0.0.0:44933", nil)
	//w.Proxy("http://127.0.0.1:8118", []string{"whatismyipaddress.com"})
	//w.Proxy("socks://127.0.0.1:7777", nil)
	//w.Proxy("socks://127.0.0.1:7777", []string{"whatismyipaddress.com"})
	w.Navigate("https://grafana.webrelay.io/playlists/play/1?kiosk")
	w.Run()
}
