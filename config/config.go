package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database connection.
var DB *gorm.DB

// LoadConfig loads the application configuration from a YAML file.
func LoadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}
}

// ConnectDB connects to the PostgreSQL database.
func ConnectDB() {
	dbConfig := viper.GetStringMapString("database") // Use GetStringMapString for type safety

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig["host"], 
		dbConfig["port"],
		dbConfig["user"], 
		dbConfig["password"], 
		dbConfig["name"],
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
}

// CloseDB closes the database connection.
func CloseDB() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Error closing database connection: %s", err)
	}
	sqlDB.Close()
}