package main

import (
	"fmt"
	"log"

	"github.com/Renal37/goph-keeper/internal/agent/client"
	"github.com/Renal37/goph-keeper/internal/agent/config"
	"github.com/Renal37/goph-keeper/internal/agent/core"
	"github.com/Renal37/goph-keeper/internal/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
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

	fmt.Println("*************************************")
	fmt.Println("Welcome GophKepeer client")
	fmt.Printf("Build version: %v \n", buildVersion)
	fmt.Printf("Build date: %v \n", buildDate)
	fmt.Println("*************************************")

	if eCfg.Command == "" {
		fmt.Println("Support command -c:")
		fmt.Println("sign-up - create new account")
		fmt.Println("sign-in - sign in with your account")
		fmt.Println("read-file - read all files on your account")
		fmt.Println("write-file - write file on your account")
		fmt.Println("delete-file - delete file from your account")
		fmt.Println("*************************************")
	}

	cl, err := client.NewClient(eCfg.ServerAddr, eCfg.Certificate, eCfg.JWT)
	if err != nil {
		lg.Sugar().Fatalf("failed create client: %s", err.Error())
	}

	err = core.Run(cl, eCfg.Command)
	if err != nil {
		lg.Sugar().Fatalf("failed command from client: %s", err.Error())
	}

	err = cl.Close()
	if err != nil {
		lg.Sugar().Fatalf("failed close client: %s", err.Error())
	}
}
