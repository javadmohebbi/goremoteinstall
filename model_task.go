package goremoteinstall

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db_dir = "/opt/yarma/goremoteinstall/var/log/tasks"

type TaskModel struct {
	Host string
	// HostID string
	Status Status
	// Desc        string
	DescPayload string
	gorm.Model
}

// TableName overrides the table name used by User to `profiles`
func (TaskModel) TableName() string {
	return "tasks"
}

func GetTaskDB(taskID string) *gorm.DB {

	_dbPath := fmt.Sprintf("%s/%s.db", _db_dir, taskID)

	if _, err := os.Stat(_dbPath); err == nil {
		log.Printf("db '%s' is exist and is will be opened", _dbPath)
	} else {
		log.Printf("db '%s' is not exist and is will be created and opened", _dbPath)
	}

	// open or create sqlite database
	db, err := gorm.Open(sqlite.Open(_dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	// check if no successful
	if err != nil {
		log.Fatalf("Can not open %v due to error: %v\n", _dbPath, err)
	}

	// migrate db
	db.AutoMigrate(&TaskModel{})

	return db

}

func (tm TaskModel) InserStatus(t *Target, status Status, payloadMsg string, db *gorm.DB) error {

	t.Status = status
	t.DescPayload = payloadMsg

	tmt := TaskModel{
		Host: tm.Host,
		// HostID: tm.HostID,
		Status: status,
		// Desc:        status.String(),
		DescPayload: payloadMsg,
	}
	if dbc := db.Create(&tmt); dbc.Error != nil {
		log.Println("db error: ", dbc.Error.Error())
		return db.Error
	}

	return nil
}
