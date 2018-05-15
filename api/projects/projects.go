package projects

import (
	"net/http"



	"github.com/masterminds/squirrel"

	"github.com/ansible-semaphore/semaphore/router"
	"github.com/ansible-semaphore/semaphore/db/models"

)

func ProjectsGetRequestHandler() func(w http.ResponseWriter, r *http.Request) {
	return router.GetGetRoute(&router.GetRequestOptions{
		Context: "user",
		NewModel: func() interface{} {
			return make([]models.AccessKey, 0)
		},
		GetQuery: router.ProjectGetCustomQueryGetter(
			"",
			[]string{"name", "type"},
			func(id string, object interface{}) squirrel.SelectBuilder {
				user := object.(*models.User)
				return squirrel.Select("p.*").
					From("project as p").
					Join("project__user as pu on pu.project_id=p.id").
					Where("pu.user_id=?", user.ID).
					OrderBy("p.name")
			}),
	},
	)
}

// GetProjects returns all projects in this users context
//func GetProjects(w http.ResponseWriter, r *http.Request) {
//	user := context.Get(r, "user").(*models.User)
//
//	query, args, err := squirrel.Select("p.*").
//		From("project as p").
//		Join("project__user as pu on pu.project_id=p.id").
//		Where("pu.user_id=?", user.ID).
//		OrderBy("p.name").
//		ToSql()
//
//	util.LogWarning(err)
//	var projects []models.Project
//	if _, err := db.Mysql.Select(&projects, query, args...); err != nil {
//		panic(err)
//	}
//
//	mulekick.WriteJSON(w, http.StatusOK, projects)
//}

// AddProject adds a new project to the database
func AddProject(w http.ResponseWriter, r *http.Request) {
	//var body models.Project
	//user := context.Get(r, "user").(*models.User)
	//
	//if err := mulekick.Bind(w, r, &body); err != nil {
	//	return
	//}

	//err := body.CreateProject()
	//if err != nil {
	//	panic(err)
	//}
	//
	//if _, err := db.Mysql.Exec("insert into project__user set project_id=?, user_id=?, `admin`=1", body.ID, user.ID); err != nil {
	//	panic(err)
	//}

	//desc := "Project Created"
	//oType := "Project"
	//if err := (models.Event{
	//	ProjectID:   &body.ID,
	//	Description: &desc,
	//	ObjectType:  &oType,
	//	ObjectID:    &body.ID,
	//	Created:     db.GetParsedTime(time.Now()),
	//}.Insert()); err != nil {
	//	panic(err)
	//}

	//mulekick.WriteJSON(w, http.StatusCreated, body)
}
