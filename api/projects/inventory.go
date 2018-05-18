package projects

import (
	"net/http"

	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"path/filepath"
	"strings"
	"os"
	"github.com/ansible-semaphore/semaphore/router"
	"github.com/ansible-semaphore/semaphore/db/models"
	"errors"
	"github.com/ansible-semaphore/semaphore/db"
)

const (
	asc  = "asc"
	desc = "desc"
)

func GetInventoryMiddleware() func(w http.ResponseWriter, r *http.Request) {
	contextKey := "project"
	identifier := "inventory"

	paramGetter := router.SimpleParamGetter(identifier)
	query := router.ProjectQueryGetter("project__" + identifier)

	return router.GetMiddleware(&router.MiddlewareOptions{
		RequestContext: contextKey,
		OutputContext:  identifier,
		QueryFunc:      query,
		ParamGetFunc:   paramGetter,
		GetObjectFunc:  func() interface{} { return new(models.Inventory) },
	},
	)
}

// InventoryMiddleware ensures an inventory exists and loads it to the context
//func InventoryMiddleware(w http.ResponseWriter, r *http.Request) {
//	project := context.Get(r, "project").(models.Project)
//	inventoryID, err := util.GetIntParam("inventory_id", w, r)
//	if err != nil {
//		return
//	}
//
//	query, args, err := squirrel.Select("*").
//		From("project__inventory").
//		Where("project_id=?", project.OutputContext).
//		Where("id=?", inventoryID).
//		ToSql()
//	util.LogWarning(err)
//
//	var inventory models.Inventory
//	if err := db.Mysql.SelectOne(&inventory, query, args...); err != nil {
//		if err == sql.ErrNoRows {
//			w.WriteHeader(http.StatusNotFound)
//			return
//		}
//
//		panic(err)
//	}
//
//	context.Set(r, "inventory", inventory)
//}

func InventoryGetRequestHandler() func(w http.ResponseWriter, r *http.Request) {
	return router.GetGetRoute(&router.GetRequestOptions{
		Context: "project",
		NewModel: func() interface{} {
			return make([]models.Environment, 0)
		},
		GetQuery: router.ProjectGetQueryGetter("inventory", []string{"name", "type"}),
	},
	)
}

// GetInventory returns an inventory from the database
//func GetInventory(w http.ResponseWriter, r *http.Request) {
//	project := context.Get(r, "project").(models.Project)
//	var inv []models.Inventory
//
//	sort := r.URL.Query().Get("sort")
//	order := r.URL.Query().Get("order")
//
//	if order != asc && order != desc {
//		order = asc
//	}
//
//	q := squirrel.Select("*").
//			From("project__inventory pi")
//
//	switch sort {
//	case "name", "type":
//		q = q.Where("pi.project_id=?", project.OutputContext).
//			OrderBy("pi." + sort + " " + order)
//	default:
//		q = q.Where("pi.project_id=?", project.OutputContext).
//		OrderBy("pi.name " + order)
//	}
//
//	query, args, err := q.ToSql()
//	util.LogWarning(err)
//
//	if _, err := db.Mysql.Select(&inv, query, args...); err != nil {
//		panic(err)
//	}
//
//	mulekick.WriteJSON(w, http.StatusOK, inv)
//}

func InventoryCreateRequestHandler() func(w http.ResponseWriter, r *http.Request) {
	return router.GetCreateRoute(&router.CreateOptions{
		Context: "project",
		NewModel: func() interface{} {
			return new(models.Inventory)
		},
		ProcessInput: inventoryValidation,
	},
		db.Mysql)
}

func inventoryValidation(context interface{}, model interface{}) error {
	inventory := model.(models.Inventory)
	switch inventory.Type {
	case "static":
		break
	case "file":
		if !IsValidInventoryPath(inventory.Inventory) {
			return errors.New("invalid inventory path")
		}
	default:
		return errors.New("unknown inventory type")
	}

	project := context.(models.Project)
	inventory.ProjectID = project.ID
	return nil
}

// AddInventory creates an inventory in the database
//func AddInventory(w http.ResponseWriter, r *http.Request) {
//	project := context.Get(r, "project").(models.Project)
//	var inventory struct {
//		Name      string `json:"name" binding:"required"`
//		KeyID     *int   `json:"key_id"`
//		SSHKeyID  int    `json:"ssh_key_id"`
//		Type      string `json:"type"`
//		Inventory string `json:"inventory"`
//	}
//
//	if err := mulekick.Bind(w, r, &inventory); err != nil {
//		return
//	}
//
//	switch inventory.Type {
//	case "static", "file":
//		break
//	default:
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	res, err := db.Mysql.Exec("insert into project__inventory set project_id=?, name=?, type=?, key_id=?, ssh_key_id=?, inventory=?", project.ID, inventory.Name, inventory.Type, inventory.KeyID, inventory.SSHKeyID, inventory.Inventory)
//	if err != nil {
//		panic(err)
//	}
//
//	insertID, err := res.LastInsertId()
//	util.LogWarning(err)
//	insertIDInt := int(insertID)
//	objType := "inventory"
//
//	desc := "Inventory " + inventory.Name + " created"
//	if err := (models.Event{
//		ProjectID:   &project.ID,
//		ObjectType:  &objType,
//		ObjectID:    &insertIDInt,
//		Description: &desc,
//	}.Insert()); err != nil {
//		panic(err)
//	}
//
//	inv := models.Inventory{
//		ID:        insertIDInt,
//		Name:      inventory.Name,
//		ProjectID: project.ID,
//		Inventory: inventory.Inventory,
//		KeyID:     inventory.KeyID,
//		SSHKeyID:  &inventory.SSHKeyID,
//		Type:      inventory.Type,
//	}
//
//	mulekick.WriteJSON(w, http.StatusCreated, inv)
//}

// IsValidInventoryPath tests a path to ensure it is below the cwd
func IsValidInventoryPath(path string) bool {

	currentPath, err := os.Getwd()
	if err != nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(currentPath, absPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relPath, "..")
}

func InventoryPutRequestHandler() func(w http.ResponseWriter, r *http.Request) {
	return router.GetPutRoute(&router.PutOptions{
		Context: "inventory",
		NewModel: func() interface{} {
			return new(models.Inventory)
		},
		ProcessInput: inventoryValidation,
	},
		db.Mysql)
}

// UpdateInventory writes updated values to an existing inventory item in the database
//func UpdateInventory(w http.ResponseWriter, r *http.Request) {
//	oldInventory := context.Get(r, "inventory").(models.Inventory)
//
//	var inventory struct {
//		Name      string `json:"name" binding:"required"`
//		KeyID     *int   `json:"key_id"`
//		SSHKeyID  int    `json:"ssh_key_id"`
//		Type      string `json:"type"`
//		Inventory string `json:"inventory"`
//	}
//
//	if err := mulekick.Bind(w, r, &inventory); err != nil {
//		return
//	}
//
//	switch inventory.Type {
//	case "static":
//		break
//	case "file":
//		if !IsValidInventoryPath(inventory.Inventory) {
//			panic("Invalid inventory path")
//		}
//	default:
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	if _, err := db.Mysql.Exec("update project__inventory set name=?, type=?, key_id=?, ssh_key_id=?, inventory=? where id=?", inventory.Name, inventory.Type, inventory.KeyID, inventory.SSHKeyID, inventory.Inventory, oldInventory.ID); err != nil {
//		panic(err)
//	}
//
//	desc := "Inventory " + inventory.Name + " updated"
//	objType := "inventory"
//	if err := (models.Event{
//		ProjectID:   &oldInventory.ProjectID,
//		Description: &desc,
//		ObjectID:    &oldInventory.ID,
//		ObjectType:  &objType,
//	}.Insert()); err != nil {
//		panic(err)
//	}
//
//	w.WriteHeader(http.StatusNoContent)
//}

// RemoveInventory deletes an inventory from the database
func RemoveInventory(w http.ResponseWriter, r *http.Request) {
	inventory := context.Get(r, "inventory").(models.Inventory)

	templatesC, err := db.Mysql.SelectInt("select count(1) from project__template where project_id=? and inventory_id=?", inventory.ProjectID, inventory.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Inventory is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := db.Mysql.Exec("update project__inventory set removed=1 where id=?", inventory.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := db.Mysql.Exec("delete from project__inventory where id=?", inventory.ID); err != nil {
		panic(err)
	}

	//desc := "Inventory " + inventory.Name + " deleted"
	//if err := (models.Event{
	//	ProjectID:   &inventory.ProjectID,
	//	Description: &desc,
	//}.Insert()); err != nil {
	//	panic(err)
	//}

	w.WriteHeader(http.StatusNoContent)
}
