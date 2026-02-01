package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/cephei8/greener/assets"
	"github.com/cephei8/greener/core"
	"github.com/cephei8/greener/core/dbutil"
	"github.com/cephei8/greener/core/mcp"
	"github.com/cephei8/greener/core/oauth"
	"github.com/cephei8/greener/core/sse"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	DatabaseURL string `env:"GREENER_DATABASE_URL"`
	AuthSecret  string `env:"GREENER_AUTH_SECRET"`
	AuthIssuer  string `env:"GREENER_AUTH_ISSUER"`
	Port        int    `env:"GREENER_PORT" envDefault:"8080"`
	Verbose     bool   `env:"GREENER_VERBOSE_OUTPUT"`
}

type Template struct {
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "template not found")
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}

	flag.StringVar(&cfg.DatabaseURL, "db-url", cfg.DatabaseURL, "Database URL")
	flag.StringVar(&cfg.AuthSecret, "auth-secret", cfg.AuthSecret, "Authentication secret key")
	flag.StringVar(&cfg.AuthIssuer, "base-url", cfg.AuthIssuer, "External base URL (only needed when behind a proxy)")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Port to listen on")
	flag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "Enable verbose output")
	flag.Parse()

	issuer := cfg.AuthIssuer
	if issuer == "" {
		issuer = fmt.Sprintf("http://localhost:%d", cfg.Port)
	}

	if cfg.DatabaseURL == "" {
		fmt.Fprintf(os.Stderr, "Error: database URL is required\n")
		fmt.Fprintf(os.Stderr, "Usage: Set GREENER_DATABASE_URL or use --db-url flag\n")
		os.Exit(1)
	}

	if cfg.AuthSecret == "" {
		fmt.Fprintf(os.Stderr, "Error: authentication secret is required\n")
		fmt.Fprintf(os.Stderr, "Usage: Set GREENER_AUTH_SECRET or use --auth-secret flag\n")
		os.Exit(1)
	}

	if err := core.Migrate(cfg.DatabaseURL, cfg.Verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}

	db, err := dbutil.Init(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	oauthServer := oauth.NewServer(db, issuer)

	sseHub := sse.NewHub()
	go sseHub.Run()

	mcpServer := mcp.NewMCPServer(db, sseHub)

	e := echo.New()

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
	}

	templates := make(map[string]*template.Template)

	componentTemplates := []string{
		"templates/base.html",
		"templates/components/navbar.html",
		"templates/components/status_icon.html",
	}

	templates["login.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/login.html")...))
	templates["testcases.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/query_editor.html", "templates/testcases.html")...))
	templates["testcase_detail.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/testcase_detail.html")...))
	templates["sessions.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/query_editor.html", "templates/sessions.html")...))
	templates["session_detail.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/session_detail.html")...))
	templates["groups.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/query_editor.html", "templates/groups.html")...))
	templates["apikeys.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/apikeys.html")...))
	templates["oauth_authorize.html"] = template.Must(template.New("").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(componentTemplates, "templates/oauth_authorize.html")...))

	tableComponents := []string{"templates/components/status_icon.html"}
	templates["testcases_table.html"] = template.Must(template.New("testcases_table.html").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(tableComponents, "templates/testcases_table.html")...))
	templates["sessions_table.html"] = template.Must(template.New("sessions_table.html").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(tableComponents, "templates/sessions_table.html")...))
	templates["groups_table.html"] = template.Must(template.New("groups_table.html").
		Funcs(funcMap).
		ParseFS(assets.TemplatesFS, append(tableComponents, "templates/groups_table.html")...))

	queryService := core.NewQueryService(db)

	e.Renderer = &Template{templates: templates}
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(cfg.AuthSecret))))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			c.Set("queryService", queryService)
			return next(c)
		}
	})

	e.GET("/static/*", echo.WrapHandler(http.FileServer(http.FS(assets.StaticFS))))

	e.GET("/", core.IndexHandler)
	e.GET("/login", core.LoginPageHandler)
	e.POST("/login", core.LoginHandler)
	e.POST("/logout", core.LogoutHandler)

	e.GET("/.well-known/oauth-authorization-server", oauthServer.MetadataHandler)
	e.GET("/oauth/authorize", oauthServer.AuthorizePageHandler)
	e.POST("/oauth/authorize", oauthServer.AuthorizeHandler)
	e.POST("/oauth/token", oauthServer.TokenHandler)
	e.POST("/oauth/register", oauthServer.RegisterHandler)

	e.GET("/testcases", core.TestcasesHandler)
	e.POST("/testcases/query", core.TestcasesHandler)
	e.GET("/testcases/:id/details", core.TestcaseDetailHandler)
	e.GET("/sessions", core.SessionsHandler)
	e.POST("/sessions/query", core.SessionsHandler)
	e.GET("/sessions/:id/details", core.SessionDetailHandler)
	e.GET("/groups", core.GroupsHandler)
	e.POST("/groups/query", core.GroupsHandler)
	e.GET("/api-keys", core.APIKeysHandler)
	e.POST("/api-keys/create", core.CreateAPIKeyHandler)
	e.DELETE("/api-keys/:id", core.DeleteAPIKeyHandler)

	apiV1 := e.Group("/api/v1")
	apiV1.GET("/sse/events", sse.NewHandler(sseHub))
	apiV1.POST("/sse/set-primary", sse.NewSetPrimaryHandler(sseHub))
	apiV1.Any("/mcp", mcpServer.EchoHandler(), oauthServer.BearerAuthMiddleware())

	ingressHandler := core.NewIngressHandler(db)
	apiV1Ingress := apiV1.Group("/ingress", core.APIKeyAuth(db))
	apiV1Ingress.POST("/sessions", ingressHandler.CreateSession)
	apiV1Ingress.POST("/testcases", ingressHandler.CreateTestcases)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.Port)))
}
