package main

import (
	"crypto/tls"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"snippetbox.ren.dev/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	debug          bool
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {

	addr := flag.String("addr", ":7080", "HTTP network address")

	dsn := flag.String("dsn", "web:539x0@tcp(localhost:3306)/snippetbox?parseTime=true", "MySQL data source name")

	debug := flag.Bool("debug", false, "Toggle debug mode")

	flag.Parse()

	loggerHandler := slog.NewTextHandler(os.Stdout, nil)

	logger := slog.New(loggerHandler)
	formDecoder := form.NewDecoder()

	db, err := useDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode

	app := &application{
		debug:          *debug,
		logger:         logger,
		formDecoder:    formDecoder,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	server := http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	logger.Info("Starting server", "addr", server.Addr)

	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	logger.Error("Failed to start server", "error", err.Error())
	os.Exit(1)
}
