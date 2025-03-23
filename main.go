package main

import (
	"log"
	"github.com/joho/godotenv"
	"coreunit.net/codec-init/internal/health"
		"coreunit.net/codec-init/internal/module"
)

var DisplayName string = "Unset"
var ShortName string = "unset"
var Version string = "?.?.?"
var Commit string = "???????"

func main() {
	log.Println(DisplayName + " - " + Version + ", build " + Commit)

	err := godotenv.Load()
	if err == nil {
		log.Println("Environment variables from dotenv-file (.env) loaded")
	} else {
		log.Println("No dotenv-file (.env) found")
	}

	userID, apiToken, apiURL := health.GetHealthEnvVar()
	standalone := userID == ""

	if !standalone  {
		log.Println("health check: starting...")
		health.InitHealthChecks(userID, apiToken, apiURL)
	}else {
		log.Println("health check: skipped")
	}

	moduleDir, systemdPath, err := module.GetModuleEnvVar()
	if err != nil {
		log.Fatal("module system while load env var: " + err.Error())
	}

	modules, err := module.LoadModules(moduleDir)
	if err != nil {
		log.Fatal("module system while load modules: " + err.Error())
	}

	err = module.ProcessModules(modules, systemdPath)
	if err != nil {
		log.Fatal("module system while process: " + err.Error())
	}
}
