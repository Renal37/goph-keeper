package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/server"
	"github.com/kuzhukin/goph-keeper/internal/server/config"
	"github.com/kuzhukin/goph-keeper/internal/yaml"
	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

const serverStopTimeout = time.Second * 30

func main() {
	defer func() {
		_ = zlog.Logger().Sync()
	}()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	zlog.Logger().Info("starting goph-keeper server...")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	configPath := flag.String("c", getenv("CONFIG_FILE", "config.yaml"), "Path config file")

	config, err := yaml.ReadYaml[config.Config](*configPath)
	if err != nil {
		return fmt.Errorf("read config, err=%w", err)
	}

	zlog.Logger().Infof("configPath: %s; config: %+v", *configPath, config)

	srvr, err := server.StartNew(config)
	if err != nil {
		return fmt.Errorf("start server err=%w", err)
	}

	sig := <-sigs
	zlog.Logger().Infof("Stop server by osSignal=%v", sig)
	if err := srvr.Stop(); err != nil {
		return fmt.Errorf("stop server, err=%w", err)
	}

	select {
	case <-srvr.WaitStop():
		zlog.Logger().Infof("Server stopped")
	case <-time.After(serverStopTimeout):
		zlog.Logger().Infof("Server stopped by timeout=%v", serverStopTimeout)
	}

	return nil
}

func getenv(name, defaultVal string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		return defaultVal
	}

	return val
}
