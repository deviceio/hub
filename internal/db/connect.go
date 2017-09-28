package db

import "github.com/jinzhu/gorm"
import _ "github.com/jinzhu/gorm/dialects/postgres"

var dbconn *gorm.DB

func Connect() {
	dbconn = gorm.Open("postgres", "host=localhost user=deviceio dbname=deviceio_hub sslmode=disable password=mypassword")
}
