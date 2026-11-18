package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"
	_ "time/tzdata" // the distroless image has no zoneinfo; calendars need time zones

	"waylight/internal/modules"
	"waylight/internal/modules/auth"
	"waylight/internal/platform/branding"
	"waylight/internal/platform/config"
	"waylight/internal/platform/crypto"
	"waylight/internal/platform/database"
	"waylight/internal/platform/demo"
	"waylight/internal/platform/httpserver"
	"waylight/internal/platform/logging"
	"waylight/internal/platform/metrics"
	"waylight/internal/platform/middleware"
	"waylight/internal/platform/router"
	"waylight/internal/platform/scheduler"
	"waylight/internal/platform/sdnotify"
	"waylight/internal/platform/spa"
	"waylight/internal/platform/webui"
	"waylight/internal/platform/wol"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

const healthcheckTimeout = 3 * time.Second

func main() {
	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	var err error
	switch cmd {
	case "serve":
		err = serve()
	case "healthcheck":
		err = healthcheck()
	case "wol-relay":
		err = wolRelay()
	case "reset-account":
		err = resetAccount()
	case "setup-status":
		err = setupStatus()
	case "version":
		fmt.Println(version)
	default:
		err = fmt.Errorf("unknown command %q (available: serve, healthcheck, wol-relay, reset-account, setup-status, version)", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// restartDelay gives the restore response time to reach the browser.
const restartDelay = 500 * time.Millisecond

// errRestart ends serve with a non-zero exit code after a restore.
var errRestart = errors.New("restarting to apply a restored backup")

// restartRequests is signalled by requestRestart (restore, #46).
var restartRequests = make(chan struct{}, 1)

// requestRestart asks serve to shut down gracefully.
func requestRestart() {
	select {
	case restartRequests <- struct{}{}:
	default:
	}
}

// app holds everything built from the configuration.
type app struct {
	cfg      config.Config
	log      *slog.Logger
	db       *sql.DB
	auth     *auth.Module
	registry *modules.Registry
}

// loadConfig reads the configuration and sets up the default logger.
func loadConfig() (config.Config, *slog.Logger, error) {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return config.Config{}, nil, fmt.Errorf("invalid configuration:\n%w", err)
	}
	log := logging.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)
	return cfg, log, nil
}

// build opens the database, constructs all enabled modules and applies their migrations.
func build(ctx context.Context, cfg config.Config, log *slog.Logger) (*app, error) {
	// A restore scheduled in the settings (#46) replaces the database
	// before it is opened.
	if applied, err := database.ApplyPendingRestore(cfg.DatabasePath); err != nil {
		return nil, err
	} else if applied {
		log.Warn("restored the database from a backup; the previous one is kept as .before-restore", "path", cfg.DatabasePath)
	}
	db, err := database.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	if cfg.DemoMode {
		if err := demo.CheckDatabase(ctx, db); err != nil {
			_ = db.Close()
			return nil, err
		}
		// The demo never reads the host; disk health and WireGuard show
		// example exports from a private directory instead.
		smart, wg, err := demo.WriteExports(filepath.Dir(cfg.DatabasePath), time.Now())
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		cfg.DiskHealthDir, cfg.WireGuardDir = smart, wg
	}
	secrets, err := crypto.NewAESGCM(cfg.SecretKey)
	if err != nil {
		return nil, err
	}
	authModule, registry, err := buildModules(cfg, log, db, secrets)
	if err != nil {
		return nil, err
	}
	for _, m := range registry.All() {
		if err := database.Migrate(ctx, db, m.Name(), m.Migrations()); err != nil {
			return nil, err
		}
	}
	if cfg.DemoMode {
		if err := demo.MarkDatabase(ctx, db); err != nil {
			return nil, err
		}
	}
	if err := applyConfigFile(ctx, cfg, log, registry); err != nil {
		return nil, err
	}
	return &app{cfg: cfg, log: log, db: db, auth: authModule, registry: registry}, nil
}

func serve() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, log, err := loadConfig()
	if err != nil {
		return err
	}
	a, err := build(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer a.db.Close()
	a.log.Info("starting Waylight", "version", version, "modules", a.registry.Names())
	sched := newScheduler(a)
	if a.cfg.DemoMode {
		reset, err := prepareDemo(ctx, a)
		if err != nil {
			return err
		}
		sched.Add(reset)
	} else if err := a.auth.AnnounceSetup(ctx); err != nil {
		return err
	}
	// A restore asks for a restart: shut down gracefully and exit with a
	// non-zero code so Docker's restart policy starts the new database.
	ctx, restart := context.WithCancel(ctx)
	defer restart()
	var restarting atomic.Bool
	go func() {
		select {
		case <-restartRequests:
			a.log.Warn("restarting to apply a restored backup")
			restarting.Store(true)
			time.Sleep(restartDelay) // let the HTTP response reach the browser
			restart()
		case <-ctx.Done():
		}
	}()
	sched.Start(ctx)
	defer sched.Wait()

	// Under systemd (native package, Raspberry Pi image) report readiness and
	// feed the watchdog while the database answers; a no-op everywhere else.
	notifier := sdnotify.FromEnv(os.Getenv, os.Getpid())
	go sdnotify.RunWatchdog(ctx, notifier, a.db.PingContext, a.log)
	if err := httpserver.Run(ctx, a.cfg.ListenAddr, newHandler(a), a.cfg.ShutdownTimeout, a.log,
		httpserver.WithReadyHook(func() { logNotifyError(a.log, notifier.Ready()) }),
		httpserver.WithStoppingHook(func() { logNotifyError(a.log, notifier.Stopping()) }),
	); err != nil {
		return err
	}
	if restarting.Load() {
		return errRestart
	}
	return nil
}

func logNotifyError(log *slog.Logger, err error) {
	if err != nil {
		log.Warn("systemd notification failed", "err", err)
	}
}

// changeSink receives the changes made through the API. It stays nil in the
// core build; the ee/ audit log sets it while the modules are built (#303).
var changeSink modules.ChangeSink

// brandingSource supplies an organisation's branding (#305); nil in a core
// build or without the ee branding module, which shows the product's.
var brandingSource branding.Source

// newHandler builds the router with all module routes and the middleware chain.
func newHandler(a *app) http.Handler {
	r := router.New(a.auth.Guard)
	managed := managedSections(a.registry)
	for _, m := range a.registry.All() {
		var mr modules.Router = r
		// The ee/ audit log (#303) sees every change to instance data and sharing.
		if changeSink != nil {
			mr = modules.ChangeRouter(mr, m, changeSink)
		}
		// Gated modules (the ee/ features, #307) check their access per request.
		if g, ok := m.(modules.Gated); ok {
			mr = modules.GatedRouter(mr, g)
		}
		// Sections owned by CONFIG_FILE refuse their config routes (#88).
		if c, ok := m.(modules.Configurable); ok && managed[c.ConfigSection()] {
			mr = modules.ManagedRouter(mr, c.ConfigRoutes())
		}
		m.RegisterRoutes(mr)
	}
	r.Protected("GET /api/modules", a.registry.Handler(frontendOnlyWidgetKinds))
	r.Handle("GET /healthz", router.HealthCheck(a.db.PingContext))
	var httpStats *metrics.HTTPStats
	if a.cfg.MetricsEnabled {
		httpStats = metrics.NewHTTPStats()
		r.Handle("GET /metrics", metricsHandler(a, httpStats))
	}
	// Frontend plugins (ADR 0021): the list is public so even the sign-in
	// page can load them; an ee build mounts its bundle here.
	r.Public("GET /api/plugins", webui.PluginsHandler(pluginURLs()))
	// Branding (#305) is public: the sign-in page shows it. The source is
	// read per request, so a module set during the build is always seen.
	source := func() branding.Source { return brandingSource }
	r.Public("GET /api/branding", branding.InfoHandler(source))
	r.Public("GET /api/branding/{kind}", branding.AssetHandler(source))
	baseManifest, _ := fs.ReadFile(webui.Dist(), "manifest.webmanifest") // empty in a build without the frontend
	r.Public("GET /manifest.webmanifest", branding.ManifestHandler(baseManifest, source))
	for _, p := range frontendPlugins {
		r.Handle("GET "+p.prefix, webui.PluginFiles(p.prefix, p.files))
	}
	r.Handle("/", spa.Handler(webui.Dist()))

	var h http.Handler = r
	if a.cfg.DemoMode {
		h = demo.Block(r)
	}
	if httpStats != nil {
		h = httpStats.Middleware(h)
	}
	return middleware.Chain(h,
		middleware.Recover(a.log),
		middleware.ClientIP(a.cfg.TrustedProxies),
		middleware.RequestLog(a.log),
		middleware.AllowList(a.cfg.AllowedCIDRs, a.log),
		middleware.SecurityHeaders,
		http.NewCrossOriginProtection().Handler,
		middleware.Compress, // JSON of the API; static files are pre-compressed (#318)
	)
}

// newScheduler registers the background jobs of all modules.
func newScheduler(a *app) *scheduler.Scheduler {
	sched := scheduler.New(a.log)
	for _, m := range a.registry.All() {
		for _, j := range m.Jobs() {
			sched.Add(scheduler.Task{Name: j.Name, Interval: j.Interval, Run: j.Run})
		}
	}
	return sched
}

func resetAccount() error {
	ctx := context.Background()
	cfg, log, err := loadConfig()
	if err != nil {
		return err
	}
	a, err := build(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer a.db.Close()
	if err := a.auth.ResetAccount(ctx); err != nil {
		return err
	}
	fmt.Println("Account removed. Restart the server and use the setup token from the log to set a new password.")
	return nil
}

// healthcheck is used by the container HEALTHCHECK, because the distroless
// image has neither curl nor wget.
func healthcheck() error {
	base, err := loopbackURL(os.Getenv("LISTEN_ADDR"))
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: healthcheckTimeout}
	res, err := client.Get(base + "/healthz") //nolint:gosec,noctx // fixed loopback URL, short-lived CLI call
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New("unhealthy: " + res.Status)
	}
	return nil
}

// wolRelay runs the Wake-on-LAN relay (compose profile "wol"). It only reads
// WOL_RELAY_LISTEN and WOL_BROADCAST and needs no database or secret key.
func wolRelay() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	env := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	return wol.Relay{
		Listen:    env("WOL_RELAY_LISTEN", ":"+wol.DefaultRelayPort),
		Broadcast: env("WOL_BROADCAST", wol.DefaultBroadcast),
		Log:       slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		Burst:     10,
		Every:     time.Second,
	}.Serve(ctx)
}
