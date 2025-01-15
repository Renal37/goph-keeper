package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Renal37/goph-keeper/internal/logger"
	repository "github.com/Renal37/goph-keeper/internal/server/adapters/repository/pg"
	"github.com/Renal37/goph-keeper/internal/server/config"
	"github.com/Renal37/goph-keeper/internal/server/core"
)

var (
	minimumCharMasterKey        = 16
	buildVersion         string = "N/A"
	buildDate            string = "N/A"
)

func main() {
	eCfg, err := config.GetConfig()
	if err != nil {
		log.Fatalln(err)
	}

	lg, err := logger.Init("info")
	if err != nil {
		log.Fatalln(err)
	}

	lg.Info(fmt.Sprintf("Build version: %v", buildVersion))
	lg.Info(fmt.Sprintf("Build date: %v", buildDate))

	if eCfg.MasterKey == "" {
		lg.Fatal("Master key not found! Please use flag -mk")
	}

	if len(eCfg.MasterKey) < minimumCharMasterKey {
		lg.Sugar().Fatalf("Minimum length master key %v characters!", minimumCharMasterKey)
	}

	repo, err := repository.NewDB(context.Background(), lg, eCfg.DSN)
	if err != nil {
		lg.Fatal(err.Error())
	}

	err = core.RunGRPCserver(lg, eCfg.Host, eCfg.CertificatePath, eCfg.CertificateKeyPath, eCfg.JWTkey, eCfg.MasterKey, repo)
	if err != nil {
		lg.Fatal(err.Error())
	}
}
