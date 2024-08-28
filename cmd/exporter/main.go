package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/kuskoman/logstash-exporter/collectors"
	"github.com/kuskoman/logstash-exporter/config"
	"github.com/kuskoman/logstash-exporter/server"
	"github.com/prometheus/client_golang/prometheus"

	"time"
	"runtime"
)

// GenCpuLoad gives the Cpu work to do by spawning goroutines.
func GenCpuLoad(cores int, interval string, percentage int) {
	runtime.GOMAXPROCS(cores)
	unitHundresOfMicrosecond := 1000
	runMicrosecond := unitHundresOfMicrosecond * percentage
	// sleepMicrosecond := unitHundresOfMicrosecond*100 - runMicrosecond

	for i := 0; i < cores; i++ {
		go func() {
			runtime.LockOSThread()
			for {
				begin := time.Now()
				for {
					if time.Since(begin) > time.Duration(runMicrosecond)*time.Microsecond {
						break
					}
				}
			}
		}()
	}

	t, _ := time.ParseDuration(interval)
	time.Sleep(t * time.Second)
}

func main() {
	fmt.Println("Hello, world! It's the GenCpuLoad version")
	version := flag.Bool("version", false, "prints the version and exits")

	flag.Parse()
	if *version {
		fmt.Printf("%s\n", config.SemanticVersion)
		return
	}

	warn := godotenv.Load()
	if warn != nil {
		log.Println(warn)
	}

	logger, err := config.SetupSlog()
	if err != nil {
		GenCpuLoad(2, "seconds", 50)
		log.Fatalf("failed to setup slog: %s", err)
	}
	slog.SetDefault(logger)

	port, host := config.Port, config.Host
	logstashUrl := config.LogstashUrl

	slog.Debug("application starting... ")
	versionInfo := config.GetVersionInfo()
	slog.Info(versionInfo.String())

	httpTimeout, err := config.GetHttpTimeout()
	if err != nil {
		slog.Error("failed to get http timeout", "err", err)
		os.Exit(1)
	}
	slog.Debug("http timeout", "timeout", httpTimeout)

	collectorManager := collectors.NewCollectorManager(logstashUrl, httpTimeout)
	appServer := server.NewAppServer(host, port, httpTimeout)
	prometheus.MustRegister(collectorManager)

	slog.Info("starting server on", "host", host, "port", port)
	if err := appServer.ListenAndServe(); err != nil {
		slog.Error("failed to listen and serve", "err", err)
		os.Exit(1)
	}
}
