package main

import "github.com/webview/webview"

func main() {

	w := webview.New(true)
	w.SetTitle("Minimal webview example")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("https://en.m.wikipedia.org/wiki/Main_Page")

	w.Run()

	defer w.Destroy()

}
