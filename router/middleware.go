package router

import (
	"database/sql"
	"github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
	"net/http"
	"github.com/ansible-semaphore/semaphore/db/models"
)

type MiddlewareOptions struct {
	RequestContext string
	OutputContext  string
	QueryFunc func(ctx interface{}, params map[string]interface{}) (string, []interface{}, error)
	ParamGetFunc func(context interface{}, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error)
	GetObjectFunc   func() interface{}
	PostRequestFunc func(w http.ResponseWriter, r *http.Request)
}

func GetMiddleware(options *MiddlewareOptions) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, options.RequestContext)
		params, err := options.ParamGetFunc(ctx, w, r)
		if err != nil {
			util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "params": params})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query, args, err := options.QueryFunc(ctx, params)
		util.LogWarningWithFields(err, logrus.Fields{"context": ctx, "params": params, "query": query, "args": args})

		data := options.GetObjectFunc()
		if err := db.Mysql.SelectOne(&data, query, args...); err != nil {
			//Only log if the error is unexpected, but always return not found
			if err != sql.ErrNoRows {
				util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "params": params, "query": query, "args": args, "data": data})
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}

		options.PostRequestFunc(w, r)
		context.Set(r, options.OutputContext, data)
	}
}

//TODO - make it work with a list of stuff
func SimpleParamGetter(paramID string) func(interface{}, http.ResponseWriter, *http.Request) (map[string]interface{}, error) {
	return func(context interface{}, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
		params := make(map[string]interface{}, 1)
		envID, err := util.GetIntParam(paramID+"_id", w, r)
		if err != nil {
			return nil, err
		}
		params[paramID] = envID
		return params, nil
	}
}

// ProjectQueryGetter returns a simple query with a single controllable where clause based on the param identifier
func ProjectQueryGetter(identifier string) func(context interface{}, params map[string]interface{}) (string, []interface{}, error) {
	return func(context interface{}, params map[string]interface{}) (string, []interface{}, error) {
		//TODO - should be a safe public method for casting to anything in a handler that writes an error and returns if it fails
		project := context.(models.Project)
		return squirrel.Select("*").
			From(identifier).
			Where("project_id=?", project.ID).
			Where("id=?", params[identifier].(string)).
			ToSql()
	}
}
