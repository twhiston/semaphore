package router

import (
	"net/http"
	"testing"
)

type TestResponseWriter struct {
	status int
	t      *testing.T
	input  []byte
	WriteCount int
}
func (w *TestResponseWriter) Header() http.Header {
	return http.Header{"test": []string{}}
}
func (w *TestResponseWriter) Write(input []byte) (int, error) {
	w.input = input
	w.t.Log(string(input))
	return w.WriteCount, nil
}
func (w *TestResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.t.Log("status_code:", statusCode)
}

type TestDb struct {
	t *testing.T
}
func (d *TestDb) Insert(object ...interface{}) error {
	d.t.Log("db Insert", object)
	return nil
}
func (d *TestDb) Update(object ...interface{}) (int64, error) {
	d.t.Log("db Update", object)
	return 1, nil
}
func (d *TestDb) Connect() error {
	return nil
}
func (d *TestDb) Init() error {
	return nil
}
func (d *TestDb) Close() {
}
