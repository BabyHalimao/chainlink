package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-plugin"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/ocr2/plugins/median"
	"github.com/smartcontractkit/chainlink/v2/plugins"
)

func main() {
	envCfg, err := plugins.GetEnvConfig()
	if err != nil {
		fmt.Printf("Failed to get environment configuration: %s\n", err)
		os.Exit(1)
	}
	lggr, closeLggr := plugins.NewLogger(envCfg)
	defer closeLggr()

	promServer := plugins.NewPromServer(envCfg.PrometheusPort(), lggr)
	err = promServer.Start()
	if err != nil {
		lggr.Fatalf("Failed to start prometheus server: %s", err)
	}
	defer func() {
		if err := promServer.Close(); err != nil {
			lggr.Warnf("Error during prometheus server shut down", err)
		}
	}()

	mp := median.NewPlugin(lggr)
	defer func() {
		logger.Sugared(lggr).ErrorIfFn(mp.Close, "pluginMedian")
	}()

	stop := make(chan struct{})
	defer close(stop)

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: loop.PluginMedianHandshakeConfig(),
		Plugins: map[string]plugin.Plugin{
			loop.PluginMedianName: &loop.GRPCPluginMedian{
				StopCh:       stop,
				Logger:       lggr,
				PluginServer: mp,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
