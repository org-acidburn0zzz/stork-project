package main

import (
	"os"
	"os/signal"
	"syscall"

	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"isc.org/stork"
	"isc.org/stork/agent"
	storkutil "isc.org/stork/util"
)

// Global Agent settings.
type AgentSettings struct {
	PrometheusOnly bool `long:"listen-prometheus-only" description:"listen only for Prometheus requests" env:"STORK_AGENT_LISTEN_PROMETHEUS_ONLY"`
	StorkOnly      bool `long:"listen-stork-only" description:"listen only for Stork Server requests" env:"STORK_AGENT_LISTEN_STORK_ONLY"`
}

func main() {
	// Setup logging
	storkutil.SetupLogging()
	log.Printf("Starting Stork Agent, version %s, build date %s", stork.Version, stork.BuildDate)

	// Start app monitor
	appMonitor := agent.NewAppMonitor()

	// Prepare agent gRPC handler
	storkAgent := agent.NewStorkAgent(appMonitor)

	// Prepare Prometheus exporters
	promKeaExporter := agent.NewPromKeaExporter(appMonitor)
	promBind9Exporter := agent.NewPromBind9Exporter(appMonitor)

	// Prepare parse for command line flags.
	var agentSettings AgentSettings
	parser := flags.NewParser(&agentSettings, flags.Default)
	parser.ShortDescription = "Stork Agent"
	parser.LongDescription = "Stork Agent"

	_, err := parser.AddGroup("Stork Server flags", "", &storkAgent.Settings)
	if err != nil {
		log.Fatalf("FATAL error: %+v", err)
	}

	_, err = parser.AddGroup("Prometheus Kea Exporter flags", "", &promKeaExporter.Settings)
	if err != nil {
		log.Fatalf("FATAL error: %+v", err)
	}

	_, err = parser.AddGroup("Prometheus BIND 9 Exporter flags", "", &promBind9Exporter.Settings)
	if err != nil {
		log.Fatalf("FATAL error: %+v", err)
	}

	if _, err := parser.Parse(); err != nil {
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		log.Fatalf("FATAL error: %+v", err)
	}

	// Only start the exporters if they're enabled.
	if !agentSettings.StorkOnly {
		promKeaExporter.Start()
		defer promKeaExporter.Shutdown()

		promBind9Exporter.Start()
		defer promBind9Exporter.Shutdown()
	}

	// Only start the agent service if it's enabled.
	if !agentSettings.PrometheusOnly {
		go storkAgent.Serve()
		defer storkAgent.Shutdown()
	}

	// We wait for ctl-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
}
