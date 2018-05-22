package db

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/util"
	_ "github.com/go-sql-driver/mysql" // imports mysql driver
	"gopkg.in/gorp.v1"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db/models"
)

type MysqlDB struct {
	gorp.DbMap
}
// Mysql is the gorp database map
// db.Connect must be called to set this up correctly
//TODO - should not be instantiated like this
var Mysql *MysqlDB

// Connect ensures that the db is connected and mapped properly with gorp
func (d *MysqlDB)Connect() error {
	db, err := connect()
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		if err = createDb(); err != nil {
			return err
		}

		db, err = connect()
		if err != nil {
			return err
		}

		if err = db.Ping(); err != nil {
			return err
		}
	}

	dc := new (MysqlDB)
	dc.Db = db
	d.Db = db
	d.Dialect = gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}
	//Mysql = &MysqlDB{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}

	return nil
}

// Init is called by main after initialization of the Mysql object to create or return an existing table map
func (d *MysqlDB) Init() error {
	d.AddTableWithName(models.APIToken{}, "user__token").SetKeys(false, "id")
	d.AddTableWithName(models.AccessKey{}, "access_key").SetKeys(true, "id")
	d.AddTableWithName(models.Environment{}, "project__environment").SetKeys(true, "id")
	d.AddTableWithName(models.Inventory{}, "project__inventory").SetKeys(true, "id")
	d.AddTableWithName(models.Project{}, "project").SetKeys(true, "id")
	d.AddTableWithName(models.Repository{}, "project__repository").SetKeys(true, "id")
	d.AddTableWithName(models.Task{}, "task").SetKeys(true, "id")
	d.AddTableWithName(models.TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
	d.AddTableWithName(models.Template{}, "project__template").SetKeys(true, "id")
	d.AddTableWithName(models.User{}, "user").SetKeys(true, "id")
	d.AddTableWithName(models.Session{}, "session").SetKeys(true, "id")
	return nil
}

// Close closes the mysql connection and reports any errors
// called from main with a defer
func (d *MysqlDB)Close() {
	err := d.Db.Close()
	if err != nil {
		log.Warn("Error closing database:" + err.Error())
	}
}

func createDb() error {
	cfg := util.Config.MySQL
	url := cfg.Username + ":" + cfg.Password + "@tcp(" + cfg.Hostname + ")/?parseTime=true&interpolateParams=true"

	db, err := sql.Open("mysql", url)
	if err != nil {
		return err
	}

	if _, err := db.Exec("create database if not exists " + cfg.DbName); err != nil {
		return err
	}

	return nil
}

func connect() (*sql.DB, error) {
	cfg := util.Config.MySQL
	url := cfg.Username + ":" + cfg.Password + "@tcp(" + cfg.Hostname + ")/" + cfg.DbName + "?parseTime=true&interpolateParams=true"

	return sql.Open("mysql", url)
}
