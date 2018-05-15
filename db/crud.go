package db

type HandlerIface interface {
	Insert(object ...interface{}) error
	Update(object ...interface{}) (int64, error)
}

type MySQLHandler struct {}
func (d *MySQLHandler)Insert(object ...interface{}) error {
	return Mysql.Insert(object)
}
func (d *MySQLHandler)Update(object ...interface{}) (int64, error) {
	return Mysql.Update(object)
}

