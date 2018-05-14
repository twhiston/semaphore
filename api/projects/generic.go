package projects

import (
	"database/sql"
	"github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"net/http"
	"github.com/masterminds/squirrel"
)

type MiddlewareOptions struct {
	contextKey    string
	ID            string
	queryFunc     func(ctx interface{}, params map[string]interface{}) (string, []interface{}, error)
	paramGetFunc  func(context interface{}, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error)
	getObjectFunc func() interface{}
}

func GetMiddleware(options MiddlewareOptions) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, options.contextKey)
		params, err := options.paramGetFunc(ctx, w, r)
		if err != nil {
			util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "params": params})
			return
		}

		query, args, err := options.queryFunc(ctx, params)
		util.LogWarningWithFields(err, logrus.Fields{"context": ctx, "params": params, "query": query, "args": args})

		data := options.getObjectFunc()
		if err := db.Mysql.SelectOne(&data, query, args...); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			util.LogErrorWithFields(err, logrus.Fields{"data": data, "query": query, "args": args})
		}

		context.Set(r, options.ID, data)
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
		project := context.(db.Project)
		return squirrel.Select("*").
			From("project__"+identifier).
			Where("project_idstring, []interface{}=?", project.ID).
			Where("id=?", params[identifier].(string)).
			ToSql()
	}
}