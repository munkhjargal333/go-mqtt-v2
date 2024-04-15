package models

import (
	"fmt"
	"log"
	"mqtt/database"
	"time"
)

type FailedLog struct {
    SerialNumber string `json:"serial_number"`
    OrgID        int    `json:"org_id"`
    BusID        int    `json:"bus_id"`
    DriverID     int    `json:"driver_id"`
    RouteID      int    `json:"route_id"`
    RouteName    string `json:"route_name"`
    StopID       int    `json:"stop_id"`
    StopName     string `json:"stop_name"`
    Cnt          int    `json:"cnt"`
    RouteAmount  int    `json:"route_amount"`
    TotalAmount  int    `json:"total_amount"`
    IsCard       bool   `json:"is_card"`
    Data         string `json:"data"`
    CardTypeID   int    `json:"card_type_id"`
    DateTime     time.Time   `json:"datetime"`
    ErrorMessages string    `json:"error_message"`
}

func (u *FailedLog) TableName() string {
	return "app_backend.failed_logs"
}

func (u *FailedLog) Migrate() error {
	db := database.DB
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.AutoMigrate(&FailedLog{}); err != nil {
		log.Println("Error migrating FailedLog:", err)
		return err
	}
	return nil
}

func (u *FailedLog) Drop() error {
	db := database.DB
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.Exec("DROP TABLE " + u.TableName()).Error; err != nil {
		log.Println("Error dropping table:", err)
		return err
	}

	return nil
}