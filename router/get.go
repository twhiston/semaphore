package router

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"net/http"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/Sirupsen/logrus"
)
type GetRequestOptions struct {
	ContextKey string
	GetObjectFunc func() interface{}
	GetQuery func(ctx interface{}, options *GetRequestOptions, opts ...interface{}) (string, []interface{}, error)
	Order string
	Sort string
}

func GetRequestHandler(options *GetRequestOptions) func(w http.ResponseWriter, r *http.Request){
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, options.ContextKey).(db.Project)
		obj := options.GetObjectFunc()

		options.Sort = r.URL.Query().Get("sort")

		options.Order = r.URL.Query().Get("order")
		if options.Order != asc && options.Order != desc {
			options.Order = asc
		}

		query, args, err := options.GetQuery(ctx, options)
		util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "query": query, "args": args})

		if _, err := db.Mysql.Select(&obj, query, args...); err != nil {
			util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "query": query, "args": args, "data": obj})
		}

		mulekick.WriteJSON(w, http.StatusOK, obj)
	}
}

func ProjectGetQueryGetter(id string, sortOptions []string) func(context interface{}, options *GetRequestOptions,  opts ...interface{}) (string, []interface{}, error) {
	return ProjectGetCustomQueryGetter(id, sortOptions, func(id string, project *db.Project)squirrel.SelectBuilder{
		return squirrel.Select("*").
			From("project__"+id+" pe").
			Where("project_id=?", project.ID)
	})
}

func ProjectGetCustomQueryGetter(id string, sortOptions []string, qb func(id string, project *db.Project)squirrel.SelectBuilder) func(context interface{}, options *GetRequestOptions,  opts ...interface{}) (string, []interface{}, error) {
	return func(context interface{}, options *GetRequestOptions, opts ...interface{}) (string, []interface{}, error) {
		project := context.(db.Project)
		q := qb(id, &project)

		if util.StringInSlice(options.Sort, sortOptions){
			q = q.Where("pe.project_id=?", project.ID).
				OrderBy("pe." + options.Sort + " " + options.Order)
		} else {
			q = q.Where("pe.project_id=?", project.ID).
				OrderBy("pe.name " + options.Order)
		}
		return q.ToSql()
	}
}

