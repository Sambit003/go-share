package config

import (
	"fmt"
	"log"

	"github.com/Sambit003/go-share/models"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Load() {
	viper.setConfigName("config")
	viper.addConfigPath(".")
	viper.setConfigType("yaml")

	if err:= viper.ReadInConfig(); err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	dbConfig := viper.GetStringMap("database")
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		&dbConfig["host"], &dbConfig["port"], &dbConfig["user"], &dbConfig["password"], &dbConfig["dbname"],
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error while connecting to database %s", err)
	}

	DB.AutoMigrate(&models.User{}, &models.File())
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Error while closing database connection %s", err)
	}
	sqlDB.Close()
}