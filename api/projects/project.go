package projects

import (
	"net/http"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/router"
	"github.com/ansible-semaphore/semaphore/db/models"
	"github.com/ansible-semaphore/semaphore/db"
)

func GetProjectMiddleware() func(w http.ResponseWriter, r *http.Request) {
	contextKey := "user"
	identifier := "project"

	paramGetter := router.SimpleParamGetter(identifier)
	query := func(context interface{}, params map[string]interface{}) (string, []interface{}, error) {
		user := context.(*models.User)
		return squirrel.Select("p.*").
			From("project as p").
			Join("project__user as pu on pu.project_id=p.id").
			Where("p.id=?", params["project"]).
			Where("pu.user_id=?", user.ID).
			ToSql()
	}

	return router.GetMiddleware(&router.MiddlewareOptions{
		RequestContext: contextKey,
		OutputContext:  identifier,
		QueryFunc:      query,
		ParamGetFunc:   paramGetter,
		GetObjectFunc:  func() interface{} { return new(models.Environment) },
	},
	)
}

// ProjectMiddleware ensures a project exists and loads it to the context
//func ProjectMiddleware(w http.ResponseWriter, r *http.Request) {
//	user := context.Get(r, "user").(*models.User)
//
//	projectID, err := util.GetIntParam("project_id", w, r)
//	if err != nil {
//		return
//	}
//
//	query, args, err := squirrel.Select("p.*").
//		From("project as p").
//		Join("project__user as pu on pu.project_id=p.id").
//		Where("p.id=?", projectID).
//		Where("pu.user_id=?", user.OutputContext).
//		ToSql()
//	util.LogWarning(err)
//
//	var project models.Project
//	if err := db.Mysql.SelectOne(&project, query, args...); err != nil {
//		if err == sql.ErrNoRows {
//			w.WriteHeader(http.StatusNotFound)
//			return
//		}
//
//		panic(err)
//	}
//
//	context.Set(r, "project", project)
//}

//GetProject returns a project details
func GetProject(w http.ResponseWriter, r *http.Request) {
	mulekick.WriteJSON(w, http.StatusOK, context.Get(r, "project"))
}

// MustBeAdmin ensures that the user has administrator rights
//TODO - move to authentication package, where we can implement the future changes to permissions
func MustBeAdmin(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	user := context.Get(r, "user").(*models.User)

	userC, err := db.Mysql.SelectInt("select count(1) from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=? and pu.admin=1", user.ID, project.ID)
	if err != nil {
		panic(err)
	}

	if userC == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

// UpdateProject saves updated project details to the database
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var body struct {
		Name      string `json:"name"`
		Alert     bool   `json:"alert"`
		AlertChat string `json:"alert_chat"`
	}

	if err := mulekick.Bind(w, r, &body); err != nil {
		return
	}

	if _, err := db.Mysql.Exec("update project set name=?, alert=?, alert_chat=? where id=?", body.Name, body.Alert, body.AlertChat, project.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteProject removes a project from the database
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)

	tx, err := db.Mysql.Begin()
	if err != nil {
		panic(err)
	}

	statements := []string{
		"delete tao from task__output as tao join task as t on t.id=tao.task_id join project__template as pt on pt.id=t.template_id where pt.project_id=?",
		"delete t from task as t join project__template as pt on pt.id=t.template_id where pt.project_id=?",
		"delete from project__template where project_id=?",
		"delete from project__user where project_id=?",
		"delete from project__repository where project_id=?",
		"delete from project__inventory where project_id=?",
		"delete from access_key where project_id=?",
		"delete from project where id=?",
	}

	for _, statement := range statements {
		_, err := tx.Exec(statement, project.ID)

		if err != nil {
			err = tx.Rollback()
			util.LogWarning(err)
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
