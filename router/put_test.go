package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func getPutRoute(t *testing.T, process func(context interface{}, model interface{}) error) func(w http.ResponseWriter, r *http.Request) {
	database := TestDb{t: t}
	return GetPutRoute(&PutOptions{
		Context:      "",
		NewModel:     func() interface{} { return map[string]string{"test": "test"} },
		ProcessInput: process,
	}, &database)
}

func TestPutRoute(t *testing.T) {

	route := getPutRoute(t, func(context interface{}, model interface{}) error {
		t.Log("Input processed")
		return nil
	})

	//now call the route
	w := TestResponseWriter{t: t}
	// Simulate no rows updated
	w.WriteCount = 0
	//arbitrary route since it never gets near the real api
	r, err := http.NewRequest("PUT", "/projects", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	route(&w, r)
	//Empty input should cause a fail
	if w.status != 400 {
		t.Fail()
	}

	r, err = http.NewRequest("PUT", "/projects", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	//Simulate a single row updated
	w.WriteCount = 1
	route(&w, r)
	//In this test case any valid data input is regarded as written
	if w.status != 204 {
		t.Fail()
	}

}

func TestFailingPutValidationPath(t *testing.T) {
	route := getPutRoute(t, func(context interface{}, model interface{}) error {
		t.Log("Test go boom now")
		return errors.New("boom")
	})

	//now call the route
	w := TestResponseWriter{t: t}
	r, err := http.NewRequest("PUT", "/projects", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	route(&w, r)
	//This should fail on the validation state
	if w.status != 400 {
		t.Fail()
	}
	body := make(map[string]string)
	err = json.Unmarshal(w.input, &body)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := body["error"]
	if !ok {
		t.Fatal("should exit due to invalid json, but no error key exists in return body")
	}
}
