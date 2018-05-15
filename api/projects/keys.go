package projects

import (
	"net/http"

	"github.com/ansible-semaphore/semaphore/router"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/db/models"
	"github.com/ansible-semaphore/semaphore/db"
)

func GetKeysMiddleware() func(w http.ResponseWriter, r *http.Request) {
	contextKey := "project"
	identifier := "key"

	paramGetter := router.SimpleParamGetter(identifier)
	query := router.ProjectQueryGetter("access_key")

	return router.GetMiddleware(&router.MiddlewareOptions{
		RequestContext: contextKey,
		OutputContext:  "accessKey",
		QueryFunc:      query,
		ParamGetFunc:   paramGetter,
		GetObjectFunc:  func() interface{} { return new(models.Environment) },
	},
	)
}

// KeyMiddleware ensures a key exists and loads it to the context
//func KeyMiddleware(w http.ResponseWriter, r *http.Request) {
//	project := context.Get(r, "project").(models.Project)
//	keyID, err := util.GetIntParam("key_id", w, r)
//	if err != nil {
//		return
//	}
//
//	var key models.AccessKey
//	if err := db.Mysql.SelectOne(&key, "select * from access_key where project_id=? and id=?", project.OutputContext, keyID); err != nil {
//		if err == sql.ErrNoRows {
//			w.WriteHeader(http.StatusNotFound)
//			return
//		}
//
//		panic(err)
//	}
//
//	context.Set(r, "accessKey", key)
//}

func KeysGetRequestHandler() func(w http.ResponseWriter, r *http.Request) {
	return router.GetGetRoute(&router.GetRequestOptions{
		Context: "project",
		NewModel: func() interface{} {
			return make([]models.AccessKey, 0)
		},
		GetQuery: router.ProjectGetCustomQueryGetter(
			"",
			[]string{"name", "type"},
			func(id string, object interface{}) squirrel.SelectBuilder {
				return squirrel.Select("t.id",
					"t.name",
					"t.type",
					"t.project_id",
					"t.key",
					"t.removed").
					From("access_key t")
			}),
	},
	)
}

// GetKeys retrieves sorted keys from the database
//func GetKeys(w http.ResponseWriter, r *http.Request) {
//	project := context.Get(r, "project").(models.Project)
//	var keys []models.AccessKey
//
//	sort := r.URL.Query().Get("sort")
//	order := r.URL.Query().Get("order")
//
//	if order != "asc" && order != "desc" {
//		order = "asc"
//	}
//
//	q := squirrel.Select("ak.id",
//		"ak.name",
//		"ak.type",
//		"ak.project_id",
//		"ak.key",
//		"ak.removed").
//		From("access_key ak")
//
//	if t := r.URL.Query().Get("type"); len(t) > 0 {
//		q = q.Where("type=?", t)
//	}
//
//	switch sort {
//	case "name", "type":
//		q = q.Where("ak.project_id=?", project.OutputContext).
//			OrderBy("ak." + sort + " " + order)
//	default:
//		q = q.Where("ak.project_id=?", project.OutputContext).
//			OrderBy("ak.name " + order)
//	}
//
//	query, args, err := q.ToSql()
//	util.LogWarning(err)
//
//	if _, err := db.Mysql.Select(&keys, query, args...); err != nil {
//		panic(err)
//	}
//
//	mulekick.WriteJSON(w, http.StatusOK, keys)
//}

// AddKey adds a new key to the database
func AddKey(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var key models.AccessKey

	if err := mulekick.Bind(w, r, &key); err != nil {
		return
	}

	switch key.Type {
	case "aws", "gcloud", "do":
		break
	case "ssh":
		if key.Secret == nil || len(*key.Secret) == 0 {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "SSH Secret empty",
			})
			return
		}
	default:
		mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid key type",
		})
		return
	}

	secret := *key.Secret + "\n"

	_, err := db.Mysql.Exec("insert into access_key set name=?, type=?, project_id=?, `key`=?, secret=?", key.Name, key.Type, project.ID, key.Key, secret)
	if err != nil {
		panic(err)
	}

	//insertID, err := res.LastInsertId()
	//util.LogWarning(err)
	//insertIDInt := int(insertID)
	//objType := "key"

	//desc := "Access Key " + key.Name + " created"
	//if err := (models.Event{
	//	ProjectID:   &project.ID,
	//	ObjectType:  &objType,
	//	ObjectID:    &insertIDInt,
	//	Description: &desc,
	//}.Insert()); err != nil {
	//	panic(err)
	//}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateKey updates key in database
// nolint: gocyclo
func UpdateKey(w http.ResponseWriter, r *http.Request) {
	var key models.AccessKey
	oldKey := context.Get(r, "accessKey").(models.AccessKey)

	if err := mulekick.Bind(w, r, &key); err != nil {
		return
	}

	switch key.Type {
	case "aws", "gcloud", "do":
		break
	case "ssh":
		if key.Secret == nil || len(*key.Secret) == 0 {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "SSH Secret empty",
			})
			return
		}
	default:
		mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid key type",
		})
		return
	}

	if key.Secret == nil || len(*key.Secret) == 0 {
		// override secret
		key.Secret = oldKey.Secret
	} else {
		secret := *key.Secret + "\n"
		key.Secret = &secret
	}

	if _, err := db.Mysql.Exec("update access_key set name=?, type=?, `key`=?, secret=? where id=?", key.Name, key.Type, key.Key, key.Secret, oldKey.ID); err != nil {
		panic(err)
	}

	//desc := "Access Key " + key.Name + " updated"
	//objType := "key"
	//if err := (models.Event{
	//	ProjectID:   oldKey.ProjectID,
	//	Description: &desc,
	//	ObjectID:    &oldKey.ID,
	//	ObjectType:  &objType,
	//}.Insert()); err != nil {
	//	panic(err)
	//}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveKey deletes a key from the database
func RemoveKey(w http.ResponseWriter, r *http.Request) {
	key := context.Get(r, "accessKey").(models.AccessKey)

	templatesC, err := db.Mysql.SelectInt("select count(1) from project__template where project_id=? and ssh_key_id=?", *key.ProjectID, key.ID)
	if err != nil {
		panic(err)
	}

	inventoryC, err := db.Mysql.SelectInt("select count(1) from project__inventory where project_id=? and ssh_key_id=?", *key.ProjectID, key.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 || inventoryC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Key is in use by one or more templates / inventory",
				"inUse": true,
			})

			return
		}

		if _, err := db.Mysql.Exec("update access_key set removed=1 where id=?", key.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := db.Mysql.Exec("delete from access_key where id=?", key.ID); err != nil {
		panic(err)
	}

	//desc := "Access Key " + key.Name + " deleted"
	//if err := (models.Event{
	//	ProjectID:   key.ProjectID,
	//	Description: &desc,
	//}.Insert()); err != nil {
	//	panic(err)
	//}

	w.WriteHeader(http.StatusNoContent)
}
