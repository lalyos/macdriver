package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	_ "embed"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"github.com/progrium/macdriver/webkit"
)

func main() {
	runtime.LockOSThread()
	var err error

	//go:embed index.html
	var defaultIndex []byte

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(defaultIndex)
	})
	http.DefaultServeMux.HandleFunc("/exit", func(w http.ResponseWriter, r *http.Request) {
		os.Exit(0)
	})

	srv := http.Server{}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	go srv.Serve(ln)

	app := cocoa.NSApp_WithDidLaunch(func(notification objc.Object) {
		config := webkit.WKWebViewConfiguration_New()
		config.Preferences().SetValueForKey(core.True, core.String("developerExtrasEnabled"))

		wv := webkit.WKWebView_Init(cocoa.NSScreen_Main().Frame(), config)
		wv.SetOpaque(false)
		wv.SetBackgroundColor(cocoa.NSColor_Clear())
		wv.SetValueForKey(core.False, core.String("drawsBackground"))

		parts := strings.Split(ln.Addr().String(), ":")
		log.Printf("to exit: curl http://127.1:%s/exit\n", parts[len(parts)-1])
		url := core.URL(fmt.Sprintf("http://%s", ln.Addr().String()))
		req := core.NSURLRequest_Init(url)
		wv.LoadRequest(req)

		w := cocoa.NSWindow_Init(cocoa.NSScreen_Main().Frame(), cocoa.NSClosableWindowMask,
			cocoa.NSBackingStoreBuffered, false)
		w.SetContentView(wv)
		w.SetBackgroundColor(cocoa.NSColor_Clear())
		w.SetOpaque(false)
		w.SetTitleVisibility(cocoa.NSWindowTitleHidden)
		w.SetTitlebarAppearsTransparent(true)
		//w.SetIgnoresMouseEvents(false)
		w.SetIgnoresMouseEvents(true)
		w.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
		w.MakeKeyAndOrderFront(w)

		events := make(chan cocoa.NSEvent)
		go func() {
			for e := range events {
				log.Println("keycode:", e.KeyCode())
				if e.KeyCode() == 100 {
					if w.IgnoresMouseEvents() {
						fmt.Println("Mouse events on")
						w.SetIgnoresMouseEvents(false)
					} else {
						fmt.Println("Mouse events off")
						w.SetIgnoresMouseEvents(true)
					}
				}
				e.Release()
			}
		}()
		cocoa.NSEvent_GlobalMonitorMatchingMask(cocoa.NSEventMaskKeyDown, events)

	})
	//app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyRegular)
	app.ActivateIgnoringOtherApps(true)

	log.Printf("topframe 0.1.0 by progrium\n")
	app.Run()
}
