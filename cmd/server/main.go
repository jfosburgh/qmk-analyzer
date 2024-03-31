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

	"github.com/qmk-analyzer/internal/qmk"
)

type config struct {
	port    int
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}
	keyboardDir string
	layoutDir   string
}

type application struct {
	cfg       config
	logger    *slog.Logger
	mux       *http.ServeMux
	wg        sync.WaitGroup
	qmkHelper *qmk.QMKHelper
	templates *template.Template
}

func main() {
	app := application{
		cfg:    config{},
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	flag.IntVar(&app.cfg.port, "port", 8080, "HTTP server port")

	flag.BoolVar(&app.cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Float64Var(&app.cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&app.cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")

	flag.StringVar(&app.cfg.keyboardDir, "keyboard-dir", "keyboards/", "Root directory for qmk keyboards")
	flag.StringVar(&app.cfg.layoutDir, "layout-dir", "layouts/", "Root directory for qmk layouts")

	flag.Parse()

	qmkHelper, err := qmk.NewQMKHelper(app.cfg.keyboardDir, app.cfg.layoutDir)
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
