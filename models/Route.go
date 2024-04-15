package models

import (
	"fmt"
	"mqtt/database"
	"time"

	"gorm.io/gorm"
)

type ImapKafka struct {
	OnlineNo  uint    `json:"onlineNo"`
	Azimuth   int     `json:"azimuth"`
	RouteCode uint    `json:"routeCode"`
	RouteName string  `json:"routeName"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Speed     int     `json:"speed"`
}
type ImapHeadKafka struct {
	OnlineNo  uint    `json:"onlineNo"`
	Azimuth   int     `json:"azimuth"`
	RouteCode uint    `json:"routeCode"`
	RouteName string  `json:"routeName"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	GetOn     int     `json:"getOn"`
	GetOff    int     `json:"getOff"`
	GPSTime   string  `json:"gpsTime"`
	Speed     int     `json:"speed"`
}
type Route struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	RouteID             uint           `json:"route_id,string"`
	RouteName           string         `json:"route_name" validate:"required"`
	RouteCode           string         `json:"route_code" validate:"required"`
	RouteAlias          string         `json:"route_alias" gorm:"type:varchar(50)"`
	RouteType           string         `json:"route_type" gorm:"type:varchar(50)"`
	FleetID             string         `json:"fleet_id" validate:"required"`
	FleetName           string         `json:"fleet_name" gorm:"type:varchar(200)" validate:"required"`
	OrganizationID      uint           `json:"organization_id" validate:"required"`
	OrganizationName    string         `json:"organization_name" gorm:"type:varchar(200)" validate:"required"`
	UpFirstDepartTime   string         `json:"up_first_depart_time" gorm:"type:varchar(50)" validate:"required"`
	UpLastDepartTime    string         `json:"up_last_depart_time" gorm:"type:varchar(50)" validate:"required"`
	DownFirstDepartTime string         `json:"down_first_depart_time" gorm:"type:varchar(50)" validate:"required"`
	DownLastDepartTime  string         `json:"down_last_depart_time" gorm:"type:varchar(50)" validate:"required"`
	UpStationCount      int            `json:"up_station_count"`
	DownStationCount    int            `json:"down_station_count"`
	FullPrice           float64        `json:"full_price" gorm:"type:numeric"`
	Price               float64        `json:"price" gorm:"type:numeric"`
	UpMile              float64        `json:"up_mile" gorm:"type:numeric"`
	DownMile            float64        `json:"down_mile" gorm:"type:numeric"`
	Status              string         `json:"status" gorm:"type:varchar(2);default:'A'"`
	ObjectVersion       uint           `json:"-" gorm:"default:1"`
	CreatedByID         uint           `json:"-"`
	CreatedAt           *time.Time     `json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedByID         uint           `json:"-"`
	UpdatedAt           *time.Time     `json:"-" gorm:"autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u *Route) TableName() string {
	return fmt.Sprintf("%s.bs_routes", "app_backend")
}

func (u *Route) Migrate() error {
	db := database.DB

	if err := db.AutoMigrate(&Route{}); err != nil {
		return err
	}

	return nil
}
