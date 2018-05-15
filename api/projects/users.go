package projects

import (
	"net/http"

	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/router"
	"github.com/ansible-semaphore/semaphore/db/models"
	"github.com/ansible-semaphore/semaphore/db"
)

func GetUsersMiddleware() func(w http.ResponseWriter, r *http.Request) {
	contextKey := "project"
	identifier := "user"

	paramGetter := router.SimpleParamGetter(identifier)
	query := func(context interface{}, params map[string]interface{}) (string, []interface{}, error) {
		project := context.(models.Project)
		return squirrel.Select("u.*").
			From("project__user as pu").
			Join("user as u on pu.user_id=u.id").
			Where("pu.project_id=?", project.ID).
			Where("pu.user_id=?", params["user"]).
			ToSql()
	}

	return router.GetMiddleware(&router.MiddlewareOptions{
		RequestContext: contextKey,
		OutputContext:  "projectUser",
		QueryFunc:      query,
		ParamGetFunc:   paramGetter,
		GetObjectFunc:  func() interface{} { return new(models.Environment) },
	},
	)
}

// UserMiddleware ensures a user exists and loads it to the context
//func UserMiddleware(w http.ResponseWriter, r *http.Request) {
//	project := context.Get(r, "project").(models.Project)
//	userID, err := util.GetIntParam("user_id", w, r)
//	if err != nil {
//		return
//	}
//
//	var user models.User
//	if err := db.Mysql.SelectOne(&user, "select u.* from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=?", userID, project.OutputContext); err != nil {
//		if err == sql.ErrNoRows {
//			w.WriteHeader(http.StatusNotFound)
//			return
//		}
//
//		panic(err)
//	}
//
//	context.Set(r, "projectUser", user)
//}

// GetUsers returns all users in a project
func GetUsers(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var users []struct {
		models.User
		Admin bool `db:"admin" json:"admin"`
	}

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != asc && order != desc {
		order = asc
	}

	q := squirrel.Select("u.*").Column("pu.admin").
		From("project__user as pu").
		LeftJoin("user as u on pu.user_id=u.id").
		Where("pu.project_id=?", project.ID)

	switch sort {
	case "name", "username", "email":
		q = q.OrderBy("u." + sort + " " + order)
	case "admin":
		q = q.OrderBy("pu." + sort + " " + order)
	default:
		q = q.OrderBy("u.name " + order)
	}

	query, args, _ := q.ToSql()

	if _, err := db.Mysql.Select(&users, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, users)
}

// AddUser adds a user to a projects team in the database
func AddUser(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var user struct {
		UserID int  `json:"user_id" binding:"required"`
		Admin  bool `json:"admin"`
	}

	if err := mulekick.Bind(w, r, &user); err != nil {
		return
	}

	//TODO - check if user already exists
	if _, err := db.Mysql.Exec("insert into project__user set user_id=?, project_id=?, `admin`=?", user.UserID, project.ID, user.Admin); err != nil {
		panic(err)
	}

	//objType := "user"
	//desc := "User OutputContext " + strconv.Itoa(user.UserID) + " added to team"
	//if err := (models.Event{
	//	ProjectID:   &project.ID,
	//	ObjectType:  &objType,
	//	ObjectID:    &user.UserID,
	//	Description: &desc,
	//}.Insert()); err != nil {
	//	panic(err)
	//}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveUser removes a user from a project team
func RemoveUser(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	user := context.Get(r, "projectUser").(models.User)

	if _, err := db.Mysql.Exec("delete from project__user where user_id=? and project_id=?", user.ID, project.ID); err != nil {
		panic(err)
	}

	//objType := "user"
	//desc := "User OutputContext " + strconv.Itoa(user.ID) + " removed from team"
	//if err := (models.Event{
	//	ProjectID:   &project.ID,
	//	ObjectType:  &objType,
	//	ObjectID:    &user.ID,
	//	Description: &desc,
	//}.Insert()); err != nil {
	//	panic(err)
	//}

	w.WriteHeader(http.StatusNoContent)
}

// MakeUserAdmin writes the admin flag to the users account
func MakeUserAdmin(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	user := context.Get(r, "projectUser").(models.User)
	admin := 1

	if r.Method == "DELETE" {
		// strip admin
		admin = 0
	}

	if _, err := db.Mysql.Exec("update project__user set `admin`=? where user_id=? and project_id=?", admin, user.ID, project.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
