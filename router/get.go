package router

import (
	"github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
	"net/http"
	"github.com/ansible-semaphore/semaphore/db/models"
)

type GetRequestOptions struct {
	Context  string
	NewModel func() interface{}
	//TODO - thought, we can abstract out the whole db layer instead of doing any queries here!
	// just pass through the vars that you need to plug into the queries and go for it
	// then reuse that stuff in dredd
	GetQuery func(ctx interface{}, options *GetRequestOptions) (string, []interface{}, error)
	Order string
	Sort  string
}

func GetGetRoute(options *GetRequestOptions) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO - sucks that this is hard coded to project, even if its the only thing we need
		// defer the cast out to a function that is called like in the other funcs if pos
		ctx := context.Get(r, options.Context).(models.Project)
		obj := options.NewModel()

		options.Sort = r.URL.Query().Get("sort")

		options.Order = r.URL.Query().Get("order")
		if options.Order != asc && options.Order != desc {
			options.Order = asc
		}

		//TODO - abstract as in create router file
		query, args, err := options.GetQuery(ctx, options)
		util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "query": query, "args": args})

		if _, err := db.Mysql.Select(&obj, query, args...); err != nil {
			util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "query": query, "args": args, "data": obj})
		}

		mulekick.WriteJSON(w, http.StatusOK, obj)
	}
}

func ProjectGetQueryGetter(id string, sortOptions []string) func(context interface{}, options *GetRequestOptions) (string, []interface{}, error) {
	return ProjectGetCustomQueryGetter(
		id,
		sortOptions,
		func(typeID string, object interface{}) squirrel.SelectBuilder {
			project := object.(*models.Project)
			return squirrel.Select("*").
				From("project__" + typeID + " t").
				Where("project_id=?", project.ID)
		})
}

// qb function should call its table t so that sorts will work here
func ProjectGetCustomQueryGetter(typeID string, sortOptions []string, qb func(id string, project interface{}) squirrel.SelectBuilder) func(context interface{}, options *GetRequestOptions) (string, []interface{}, error) {
	return func(context interface{}, options *GetRequestOptions) (string, []interface{}, error) {
		project := context.(models.Project)
		q := qb(typeID, &project)

		//TODO - break out into own function that can be replaced
		if util.StringInSlice(options.Sort, sortOptions) {
			q = q.Where("t.project_id=?", project.ID).
				OrderBy("t." + options.Sort + " " + options.Order)
		} else {
			q = q.Where("t.project_id=?", project.ID).
				OrderBy("t.name " + options.Order)
		}
		return q.ToSql()
	}

}
