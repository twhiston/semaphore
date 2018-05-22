package db

type DbIface interface {
	Connect() error
	Init() error
	Close()
	Insert(object ...interface{}) error
	Update(object ...interface{}) (int64, error)
}
