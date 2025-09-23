package main

import (
	"log"
	"net/http"

	"github.com/endingwithali/2025censys/cmd/config"
	"github.com/endingwithali/2025censys/internal/api"
	"github.com/endingwithali/2025censys/internal/repo"
	"github.com/endingwithali/2025censys/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// type BackendEnv struct {
// 	DB *gorm.DB
// }

func main() {

	serverConfig := config.Load()

	db, err := gorm.Open(postgres.Open(serverConfig.DBConfig.Connection_String), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	// Setting up layers
	snapshotRepo := repo.NewSnapshotRepo(db)
	snapshotService := service.NewSnapshotService(snapshotRepo, serverConfig.HostFileConfig.Location)
	differenceSerive := service.NewDifferencesServicet()
	router := api.New(snapshotService, differenceSerive, serverConfig.HostFileConfig.MaxSize)

	log.Printf("Listening on Port %s", serverConfig.Port)
	if err = http.ListenAndServe(serverConfig.Port, router); err != nil {
		log.Fatal(err)
	}
}
