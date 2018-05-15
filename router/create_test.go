package router

import (
	"testing"
	"net/http"
	"strings"
	"errors"
	"encoding/json"
)

type TestResponseWriter struct{
	status int
	t *testing.T
	input []byte
}

func (w *TestResponseWriter)Header() http.Header{
	return http.Header{"test":[]string{}}
}

func (w *TestResponseWriter)Write(input []byte) (int, error){
	w.input = input
	w.t.Log(string(input))
	return 0, nil
}
func (w *TestResponseWriter)WriteHeader(statusCode int){
	w.status = statusCode
	w.t.Log("status_code:", statusCode)
}

type TestDbHandler struct{
	t *testing.T
}

func (d *TestDbHandler)Insert(object ...interface{}) error {
	d.t.Log("db Insert", object)
	return nil
}

func (d *TestDbHandler)Update(object ...interface{}) (int64, error) {
	d.t.Log("db Update", object)
	return 0, nil
}

func TestGetCreateRoute(t *testing.T) {

	handler := TestDbHandler{t:t}
	route := GetCreateRoute(&CreateOptions{
		Context: "",
		NewModel:     func() interface{}{return map[string]string{"test": "test"}},
		ProcessInput: func(context interface{}, model interface{}) error {
			t.Log("Input processed")
			return nil
		},
		handler: &handler,
	})

	//now call the route
	w := TestResponseWriter{t: t}
	//arbitrary route since it never gets near the real api
	r, err := http.NewRequest("GET","/projects", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	route(&w,r)
	//Empty input should cause a fail
	if w.status != 400 {
		t.Fail()
	}

	r, err = http.NewRequest("GET","/projects", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	route(&w,r)
	//In this test case any valid data input is regarded as written
	if w.status != 201 {
		t.Fail()
	}

}

func TestFailingValidationPath(t *testing.T) {
	handler := TestDbHandler{t:t}
	route := GetCreateRoute(&CreateOptions{
		Context: "",
		NewModel:     func() interface{}{return map[string]string{"test": "test"}},
		ProcessInput: func(context interface{}, model interface{}) error {
			return errors.New("boom")
		},
		handler: &handler,
	})

	//now call the route
	w := TestResponseWriter{t: t}

	r, err := http.NewRequest("GET","/projects", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	route(&w,r)
	//This should fail on the validation state
	if w.status != 400 {
		t.Fail()
	}
	body := make(map[string]string)
	err = json.Unmarshal(w.input, &body)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := body["error"]; if !ok {
		t.Fatal("should exit due to invalid json error but no error key exists")
	}
}