package database

import (
	"log"

	"github.com/fitnis/shared/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the database connection and performs auto-migration.
func InitDB() {
	var err error

	DB, err = gorm.Open(sqlite.Open("fitnis.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	log.Println("Running database migrations...")
	err = DB.AutoMigrate(
		&models.Patient{},
		&models.Examination{},
		&models.Sample{},
		&models.Prescription{},
		&models.Referral{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed.")
	log.Println("Database connection established.")
}
