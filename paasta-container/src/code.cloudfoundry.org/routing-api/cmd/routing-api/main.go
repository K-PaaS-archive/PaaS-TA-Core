package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/locket/lock"
	"code.cloudfoundry.org/routing-api"
	"code.cloudfoundry.org/routing-api/config"
	"code.cloudfoundry.org/routing-api/db"
	"code.cloudfoundry.org/routing-api/handlers"
	"code.cloudfoundry.org/routing-api/helpers"
	"code.cloudfoundry.org/routing-api/metrics"
	"code.cloudfoundry.org/routing-api/migration"
	"code.cloudfoundry.org/routing-api/models"
	uaaclient "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/cloudfoundry/dropsonde"
	"github.com/nu7hatch/gouuid"

	"code.cloudfoundry.org/clock"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"

	"github.com/tedsuo/rata"

	locketmodels "code.cloudfoundry.org/locket/models"
)

const (
	DEFAULT_ETCD_WORKERS = 25
	routingApiLockPath   = "routing_api_lock"
	sessionName          = "routing_api"
	pruningInterval      = 10 * time.Second
)

var port = flag.Uint("port", 8080, "Port to run rounting-api server on")
var configPath = flag.String("config", "", "Configuration for routing-api")
var devMode = flag.Bool("devMode", false, "Disable authentication for easier development iteration")
var ip = flag.String("ip", "", "The public ip of the routing api")

func route(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}

func main() {
	lagerflags.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, reconfigurableSink := lagerflags.New("routing-api")

	err := checkFlags()
	if err != nil {
		logger.Error("failed-check-flags", err)
		os.Exit(1)
	}

	cfg, err := config.NewConfigFromFile(*configPath, *devMode)
	if err != nil {
		logger.Error("failed-load-config", err)
		os.Exit(1)
	}

	err = dropsonde.Initialize(cfg.MetronConfig.Address+":"+cfg.MetronConfig.Port, cfg.LogGuid)
	if err != nil {
		logger.Error("failed-initialize-dropsonde", err)
		os.Exit(1)
	}

	if cfg.DebugAddress != "" {
		_, err := debugserver.Run(cfg.DebugAddress, reconfigurableSink)
		if err != nil {
			logger.Error("failed-debug-server", err, lager.Data{"debug_address": cfg.DebugAddress})
		}
	}

	database, err := buildDatabase(logger, cfg)
	if err != nil {
		os.Exit(1)
	}

	prefix := "routing_api"
	statsdClient, err := statsd.NewBufferedClient(cfg.StatsdEndpoint, prefix, cfg.StatsdClientFlushInterval, 512)
	if err != nil {
		logger.Error("failed-to-create-statsd-client", err)
		os.Exit(1)
	}
	defer func() {
		err := statsdClient.Close()
		if err != nil {
			logger.Error("failed-to-close-statsd-client", err)
		}
	}()

	apiServer := constructApiServer(cfg, database, statsdClient, logger.Session("api-server"))
	stopper := constructStopper(database)

	routerRegister := constructRouteRegister(
		cfg.LogGuid,
		cfg.SystemDomain,
		cfg.MaxTTL,
		database,
		logger.Session("route-register"),
	)
	clock := clock.NewClock()

	etcdDone := make(chan struct{})
	releaseLock := make(chan os.Signal)
	lockErrChan := make(chan error)
	metricsTicker := time.NewTicker(cfg.MetricsReportingInterval)
	metricsReporter := metrics.NewMetricsReporter(database, statsdClient, metricsTicker, logger.Session("metrics"))
	migrationProcess := runMigration(cfg, database, &cfg.Etcd, etcdDone, logger.Session("migration"))
	routerGroupSeeder := seedRouterGroups(cfg, database, logger.Session("seeding"))

	locks := grouper.Members{}

	if !cfg.SkipConsulLock {
		lockMaintainer := initializeLockMaintainer(logger, cfg.ConsulCluster.Servers, sessionName,
			cfg.ConsulCluster.LockTTL, cfg.ConsulCluster.RetryInterval, clock)
		locks = append(locks, grouper.Member{Name: "lock-maintainer", Runner: lockMaintainer})
	}

	var locketClient locketmodels.LocketClient
	if cfg.Locket.LocketAddress != "" {
		locketClient, err = locket.NewClient(logger, cfg.Locket)
		if err != nil {
			logger.Fatal("failed-to-create-locket-client", err)
		}

		lockIdentifier := &locketmodels.Resource{
			Key:   routingApiLockPath,
			Owner: cfg.UUID,
			Type:  locketmodels.LockType,
		}

		locks = append(locks, grouper.Member{Name: "sql-lock", Runner: lock.NewLockRunner(
			logger,
			locketClient,
			lockIdentifier,
			locket.DefaultSessionTTLInSeconds,
			clock,
			locket.SQLRetryInterval,
		)})
	}

	if len(locks) == 0 {
		logger.Fatal("no-locks-configured", errors.New("Lock configuration must be provided"))
	}

	lockGroup := grouper.NewOrdered(os.Interrupt, locks)
	lockAcquirer := initializeLockAcquirer(lockGroup, releaseLock, lockErrChan)
	lockReleaser := initializeLockReleaser(releaseLock, lockErrChan, cfg.ConsulCluster.RetryInterval)

	members := grouper.Members{
		grouper.Member{Name: "migration", Runner: migrationProcess},
		grouper.Member{Name: "lock-acquirer", Runner: lockAcquirer},
		grouper.Member{Name: "seed-router-groups", Runner: routerGroupSeeder},
		grouper.Member{Name: "api-server", Runner: apiServer},
		grouper.Member{Name: "conn-stopper", Runner: stopper},
		grouper.Member{Name: "route-register", Runner: routerRegister},
		grouper.Member{Name: "metrics", Runner: metricsReporter},
	}

	if isSql(cfg.SqlDB) {
		routePruner := runCleanupRoutes(database, logger)
		members = append(members, grouper.Member{Name: "sql-route-pruner", Runner: routePruner})
	}
	members = append(members, grouper.Member{Name: "lock-releaser", Runner: lockReleaser})

	group := grouper.NewOrdered(os.Interrupt, members)
	process := ifrit.Invoke(sigmon.New(group))

	// This is used by testrunner to signal ready for tests.
	logger.Info("started", lager.Data{"port": *port})

	errChan := process.Wait()
	err = <-errChan
	if err != nil {
		logger.Error("shutdown-error", err)
		os.Exit(1)
	}
	logger.Info("exited")
}

func isSql(sqlDB config.SqlDB) bool {
	return (sqlDB.Host != "" && sqlDB.Port > 0 && sqlDB.Schema != "")
}

func constructStopper(database db.DB) ifrit.Runner {
	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		close(ready)
		select {
		case <-signals:
			database.CancelWatches()
		}

		return nil
	})
}

func seedRouterGroups(cfg config.Config, database db.DB, logger lager.Logger) ifrit.Runner {
	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		if len(cfg.RouterGroups) > 0 {
			routerGroups, _ := database.ReadRouterGroups()
			// if config not empty and db is empty, seed
			if len(routerGroups) == 0 {
				for _, rg := range cfg.RouterGroups {
					guid, err := uuid.NewV4()
					if err != nil {
						logger.Error("failed to generate a guid for router group", err)
						return err
					}
					rg.Guid = guid.String()
					logger.Info("seeding", lager.Data{"router-group": rg})
					err = database.SaveRouterGroup(rg)
					if err != nil {
						logger.Error("failed to save router group from config", err)
						return err
					}
				}
			}
		}
		close(ready)
		select {
		case sig := <-signals:
			logger.Info("received-signal", lager.Data{"signal": sig})
		}
		return nil
	})
}

func runCleanupRoutes(sqlDatabase db.DB, logger lager.Logger) ifrit.Runner {
	pruneLogger := logger.Session("prune-routes")
	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		sqlDB, ok := sqlDatabase.(*db.SqlDB)
		if !ok {
			return nil
		}
		close(ready)

		sqlDB.CleanupRoutes(pruneLogger, pruningInterval, signals)
		return nil
	})
}

func runMigration(cfg config.Config, database db.DB, etcdCfg *config.Etcd, etcdDone chan struct{}, logger lager.Logger) ifrit.Runner {
	if sqlDB, ok := database.(*db.SqlDB); ok {
		return migration.NewRunner(etcdCfg, etcdDone, sqlDB, logger)
	}
	return migration.NewRunner(etcdCfg, etcdDone, nil, logger)
}

func constructRouteRegister(
	logGuid string,
	systemDomain string,
	maxTTL time.Duration,
	database db.DB,
	logger lager.Logger,
) ifrit.Runner {
	host := fmt.Sprintf("api.%s/routing", systemDomain)
	route := models.NewRoute(host, uint16(*port), *ip, logGuid, "", int(maxTTL.Seconds()))

	registerInterval := int(maxTTL.Seconds()) / 2
	ticker := time.NewTicker(time.Duration(registerInterval) * time.Second)

	return helpers.NewRouteRegister(database, route, ticker, logger)
}

func constructApiServer(cfg config.Config, database db.DB, statsdClient statsd.Statter, logger lager.Logger) ifrit.Runner {

	uaaClient, err := newUaaClient(logger, cfg)
	if err != nil {
		logger.Error("Failed to create uaa client", err)
		os.Exit(1)
	}

	issuer, err := uaaClient.FetchIssuer()
	if err != nil {
		logger.Error("Failed to get issuer configuration from UAA", err)
		os.Exit(1)
	}
	logger.Info("received-issuer", lager.Data{"issuer": issuer})

	_, err = uaaClient.FetchKey()
	if err != nil {
		logger.Error("Failed to get verification key from UAA", err)
		os.Exit(1)
	}

	validator := handlers.NewValidator()
	routesHandler := handlers.NewRoutesHandler(uaaClient, int(cfg.MaxTTL.Seconds()), validator, database, logger)
	eventStreamHandler := handlers.NewEventStreamHandler(uaaClient, database, logger, statsdClient)
	routerGroupsHandler := handlers.NewRouteGroupsHandler(uaaClient, logger, database)
	tcpMappingsHandler := handlers.NewTcpRouteMappingsHandler(uaaClient, validator, database, int(cfg.MaxTTL.Seconds()), logger)

	actions := rata.Handlers{
		routing_api.UpsertRoute:           route(routesHandler.Upsert),
		routing_api.DeleteRoute:           route(routesHandler.Delete),
		routing_api.ListRoute:             route(routesHandler.List),
		routing_api.EventStreamRoute:      route(eventStreamHandler.EventStream),
		routing_api.ListRouterGroups:      route(routerGroupsHandler.ListRouterGroups),
		routing_api.CreateRouterGroup:     route(routerGroupsHandler.CreateRouterGroup),
		routing_api.UpdateRouterGroup:     route(routerGroupsHandler.UpdateRouterGroup),
		routing_api.UpsertTcpRouteMapping: route(tcpMappingsHandler.Upsert),
		routing_api.DeleteTcpRouteMapping: route(tcpMappingsHandler.Delete),
		routing_api.ListTcpRouteMapping:   route(tcpMappingsHandler.List),
		routing_api.EventStreamTcpRoute:   route(eventStreamHandler.TcpEventStream),
	}

	handler, err := rata.NewRouter(routing_api.Routes(), actions)
	if err != nil {
		logger.Error("failed to create router", err)
		os.Exit(1)
	}

	handler = handlers.LogWrap(handler, logger)
	return http_server.New(":"+strconv.Itoa(int(*port)), handler)
}

func newUaaClient(logger lager.Logger, routingApiConfig config.Config) (uaaclient.Client, error) {
	if *devMode {
		return uaaclient.NewNoOpUaaClient(), nil
	}

	if routingApiConfig.OAuth.Port == -1 {
		logger.Fatal("tls-not-enabled", errors.New("GoRouter requires TLS enabled to get OAuth token"), lager.Data{"token-endpoint": routingApiConfig.OAuth.TokenEndpoint, "port": routingApiConfig.OAuth.Port})
	}

	scheme := "https"
	tokenURL := fmt.Sprintf("%s://%s:%d", scheme, routingApiConfig.OAuth.TokenEndpoint, routingApiConfig.OAuth.Port)

	cfg := &uaaconfig.Config{
		UaaEndpoint:      tokenURL,
		SkipVerification: routingApiConfig.OAuth.SkipSSLValidation,
		CACerts:          routingApiConfig.OAuth.CACerts,
	}
	return uaaclient.NewClient(logger, cfg, clock.NewClock())
}

func checkFlags() error {
	if *configPath == "" {
		return errors.New("No configuration file provided")
	}

	if *ip == "" {
		return errors.New("No ip address provided")
	}

	if *port > 65535 {
		return errors.New("Port must be in range 0 - 65535")
	}

	return nil
}

func initializeLockMaintainer(
	logger lager.Logger,
	consulCluster, sessionName string,
	lockTTL, lockRetryInterval time.Duration,
	clock clock.Clock,
) ifrit.Runner {
	client, err := consuladapter.NewClientFromUrl(consulCluster)
	if err != nil {
		logger.Fatal("new-client-failed", err)
	}

	return newLockRunner(logger, client, clock, lockRetryInterval, lockTTL)
}

func initializeLockAcquirer(lockRunner ifrit.Runner, releaseLock chan os.Signal, lockErrChan chan error) ifrit.Runner {
	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {

		go func() {
			var err error
			err = lockRunner.Run(releaseLock, ready)
			lockErrChan <- err
		}()

		<-signals
		return nil
	})
}

func initializeLockReleaser(releaseLock chan os.Signal, lockErrChan chan error, retryInterval time.Duration) ifrit.Runner {
	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		close(ready)
		var err error
		for {
			select {
			case signal := <-signals:
				releaseLock <- signal

			case err = <-lockErrChan:
				consulLockRetryInterval := 5 * time.Second // DefaultLockRetryTime from consul API
				//Give other routing-api enough time to grab the lock
				time.Sleep(retryInterval + consulLockRetryInterval)

				return err
			}
		}
	})
}

func newLockRunner(
	logger lager.Logger,
	consulClient consuladapter.Client,
	clock clock.Clock,
	lockRetryInterval time.Duration,
	lockTTL time.Duration,
) ifrit.Runner {
	lockSchemaPath := locket.LockSchemaPath(routingApiLockPath)

	routingApiUUID, err := uuid.NewV4()
	if err != nil {
		logger.Fatal("Couldn't generate tcp Emitter UUID", err)
	}
	routingApiID := []byte(routingApiUUID.String())

	return locket.NewLock(logger, consulClient, lockSchemaPath,
		routingApiID, clock, lockRetryInterval, lockTTL)
}

func buildDatabase(
	logger lager.Logger,
	cfg config.Config,
) (db.DB, error) {
	var err error
	var database db.DB

	// Use SQL database if available, otherwise use ETCD
	if isSql(cfg.SqlDB) {
		data := lager.Data{"host": cfg.SqlDB.Host, "port": cfg.SqlDB.Port}
		logger.Info("database", data)
		database, err = db.NewSqlDB(&cfg.SqlDB)
		if err != nil {
			logger.Error("failed-initialize-sql-connection", err, data)
			return nil, err
		}
	} else {
		logger.Info("database", lager.Data{"etcd-addresses": cfg.Etcd.NodeURLS})
		database, err = db.NewETCD(&cfg.Etcd)
		if err != nil {
			logger.Error("failed-initialize-etcd-db", err)
			return nil, err
		}
	}

	return database, nil
}
