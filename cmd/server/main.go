package main

import (
	"expvar"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/qmk-analyzer/internal/cache"
	"github.com/qmk-analyzer/internal/qmk"
)

type config struct {
	port    int
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}
	layoutDir         string
	keymapDir         string
	fingermapDir      string
	saveKeymapUploads bool
}

type application struct {
	cfg          config
	logger       *slog.Logger
	mux          *http.ServeMux
	wg           sync.WaitGroup
	qmkHelper    *qmk.QMKHelper
	templates    *template.Template
	sessionCache cache.MemoryCache
}

func main() {
	app := application{
		cfg:          config{},
		logger:       slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})),
		sessionCache: cache.NewMemoryCache(),
	}

	flag.IntVar(&app.cfg.port, "port", 8080, "HTTP server port")

	flag.BoolVar(&app.cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Float64Var(&app.cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&app.cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")

	// flag.StringVar(&app.cfg.keyboardDir, "keyboard-dir", "keyboards/", "Root directory for qmk keyboards")
	flag.StringVar(&app.cfg.layoutDir, "layout-dir", "assets/example_configs/layouts/", "Root directory for qmk layouts")
	flag.StringVar(&app.cfg.fingermapDir, "fingermap-dir", "assets/example_configs/fingermaps/", "Root directory for qmk keycodes")
	flag.StringVar(&app.cfg.keymapDir, "keymap-dir", "assets/example_configs/keymaps/", "Root directory for uploaded qmk keycodes")

	flag.BoolVar(&app.cfg.saveKeymapUploads, "save-uploads", true, "Save keymap uploads to dist")

	flag.Parse()

	qmkHelper, err := qmk.NewQMKHelper(app.cfg.layoutDir, app.cfg.keymapDir, app.cfg.fingermapDir)
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}

	app.qmkHelper = qmkHelper
	app.templates = app.parseTemplates()

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	err = app.serve()
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}
}
