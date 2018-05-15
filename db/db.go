package db

type DbIface interface {
	Connect() error
	Close()
	GetHandler() HandlerIface
}

type Db struct {
	handler HandlerIface
}

var Connection DbIface
