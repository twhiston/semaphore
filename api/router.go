package api

import (
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/projects"
	"github.com/ansible-semaphore/semaphore/sockets"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

var publicAssets = packr.NewBox("../web/public")

//JSONMiddleware ensures that all the routes respond with Json, this is added by default to all routes
func JSONMiddleware(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
}

//PlainTextMiddleware resets headers to Plain Text if needed
func PlainTextMiddleware(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain; charset=utf-8")
}

// Route declares all routes
// TODO - make this in a way that routes dont need to be manually declared
// wrap up all the api changes and the route config into type and just process them here
func Route() mulekick.Router {
	r := mulekick.New(mux.NewRouter(), mulekick.CorsMiddleware, JSONMiddleware)
	r.NotFoundHandler = http.HandlerFunc(servePublic)

	r.Get("/api/ping", PlainTextMiddleware, mulekick.PongHandler)

	// set up the namespace
	api := r.Group("/api")

	func(api mulekick.Router) {
		api.Post("/login", login)
		api.Post("/logout", logout)
	}(api.Group("/auth"))

	api.Use(authentication)

	api.Get("/ws", sockets.Handler)

	api.Get("/info", getSystemInfo)
	api.Get("/upgrade", checkUpgrade)
	api.Post("/upgrade", doUpgrade)

	func(api mulekick.Router) {
		api.Get("", getUser)
		// api.PUT("/user", misc.UpdateUser)

		api.Get("/tokens", getAPITokens)
		api.Post("/tokens", createAPIToken)
		api.Delete("/tokens/{token_id}", expireAPIToken)
	}(api.Group("/user"))

	api.Get("/projects", projects.ProjectsGetRequestHandler())
	api.Post("/projects", projects.AddProject)
	api.Get("/events", getAllEvents)
	api.Get("/events/last", getLastEvents)

	api.Get("/users", getUsers)
	api.Post("/users", addUser)
	api.Get("/users/{user_id}", getUserMiddleware, getUser)
	api.Put("/users/{user_id}", getUserMiddleware, updateUser)
	api.Post("/users/{user_id}/password", getUserMiddleware, updateUserPassword)
	api.Delete("/users/{user_id}", getUserMiddleware, deleteUser)

	func(api mulekick.Router) {
		api.Use(projects.GetProjectMiddleware())

		api.Get("", projects.GetProject)
		api.Put("", projects.MustBeAdmin, projects.UpdateProject)
		api.Delete("", projects.MustBeAdmin, projects.DeleteProject)

		api.Get("/events", getAllEvents)
		api.Get("/events/last", getLastEvents)

		api.Get("/users", projects.GetUsers)
		api.Post("/users", projects.MustBeAdmin, projects.AddUser)
		api.Post("/users/{user_id}/admin", projects.MustBeAdmin, projects.GetUsersMiddleware(), projects.MakeUserAdmin)
		api.Delete("/users/{user_id}/admin", projects.MustBeAdmin, projects.GetUsersMiddleware(), projects.MakeUserAdmin)
		api.Delete("/users/{user_id}", projects.MustBeAdmin, projects.GetUsersMiddleware(), projects.RemoveUser)

		api.Get("/keys", projects.KeysGetRequestHandler())
		api.Post("/keys", projects.AddKey)
		api.Put("/keys/{key_id}", projects.GetKeysMiddleware(), projects.UpdateKey)
		api.Delete("/keys/{key_id}", projects.GetKeysMiddleware(), projects.RemoveKey)

		api.Get("/repositories", projects.GetRepositories)
		api.Post("/repositories", projects.AddRepository)
		api.Put("/repositories/{repository_id}", projects.GetRepositoryMiddleware(), projects.UpdateRepository)
		api.Delete("/repositories/{repository_id}", projects.GetRepositoryMiddleware(), projects.RemoveRepository)

		api.Get("/inventory", projects.InventoryGetRequestHandler())
		api.Post("/inventory", projects.InventoryCreateRequestHandler())
		api.Put("/inventory/{inventory_id}", projects.GetInventoryMiddleware(), projects.InventoryPutRequestHandler())
		api.Delete("/inventory/{inventory_id}", projects.GetInventoryMiddleware(), projects.RemoveInventory)

		api.Get("/environment", projects.EnvironmentGetRequestHandler())
		api.Post("/environment", projects.EnvironmentCreateRequestHandler())
		api.Put("/environment/{environment_id}", projects.GetEnvironmentMiddleware(), projects.EnvironmentPutRequestHandler())
		api.Delete("/environment/{environment_id}", projects.GetEnvironmentMiddleware(), projects.RemoveEnvironment)

		api.Get("/templates", projects.GetTemplates)
		api.Post("/templates", projects.AddTemplate)
		api.Put("/templates/{template_id}", projects.GetTemplatesMiddleware(), projects.UpdateTemplate)
		api.Delete("/templates/{template_id}", projects.GetTemplatesMiddleware(), projects.RemoveTemplate)

		api.Get("/tasks", projects.GetAllTasks)
		api.Get("/tasks/last", projects.GetLastTasks)
		api.Post("/tasks", projects.AddTask)
		api.Get("/tasks/{task_id}/output", projects.GetTaskMiddleware(), projects.GetTaskOutput)
		api.Get("/tasks/{task_id}", projects.GetTaskMiddleware(), projects.GetTask)
		api.Delete("/tasks/{task_id}", projects.GetTaskMiddleware(), projects.RemoveTask)
	}(api.Group("/project/{project_id}"))

	return r
}

//nolint: gocyclo
func servePublic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/api") {
		mulekick.NotFoundHandler(w, r)
		return
	}

	if !strings.HasPrefix(path, "/public") {
		if len(strings.Split(path, ".")) > 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path = "/html/index.html"
	}

	path = strings.Replace(path, "/public/", "", 1)
	split := strings.Split(path, ".")
	suffix := split[len(split)-1]

	res, err := publicAssets.MustBytes(path)
	if err != nil {
		mulekick.NotFoundHandler(w, r)
		return
	}

	// replace base path
	if util.WebHostURL != nil && path == "html/index.html" {
		res = []byte(strings.Replace(string(res),
			"<base href=\"/\">",
			"<base href=\""+util.WebHostURL.String()+"\">",
			1))
	}

	contentType := "text/plain"
	switch suffix {
	case "png":
		contentType = "image/png"
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "gif":
		contentType = "image/gif"
	case "js":
		contentType = "application/javascript"
	case "css":
		contentType = "text/css"
	case "woff":
		contentType = "application/x-font-woff"
	case "ttf":
		contentType = "application/x-font-ttf"
	case "otf":
		contentType = "application/x-font-otf"
	case "html":
		contentType = "text/html"
	}

	w.Header().Set("content-type", contentType)
	_, err = w.Write(res)
	util.LogWarning(err)
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	body := map[string]interface{}{
		"version": util.Version,
		"update":  util.UpdateAvailable,
		"config": map[string]string{
			"dbHost":  util.Config.MySQL.Hostname,
			"dbName":  util.Config.MySQL.DbName,
			"dbUser":  util.Config.MySQL.Username,
			"path":    util.Config.TmpPath,
			"cmdPath": util.FindSemaphore(),
		},
	}

	if util.UpdateAvailable != nil {
		body["updateBody"] = string(blackfriday.MarkdownCommon([]byte(*util.UpdateAvailable.Body)))
	}

	mulekick.WriteJSON(w, http.StatusOK, body)
}

func checkUpgrade(w http.ResponseWriter, r *http.Request) {
	if err := util.CheckUpdate(util.Version); err != nil {
		mulekick.WriteJSON(w, 500, err)
		return
	}

	if util.UpdateAvailable != nil {
		getSystemInfo(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func doUpgrade(w http.ResponseWriter, r *http.Request) {
	util.LogError(util.DoUpgrade(util.Version))
}
