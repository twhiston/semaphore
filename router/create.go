package router

import (
	"github.com/castawaylabs/mulekick"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"
	"github.com/gorilla/context"
	"github.com/ansible-semaphore/semaphore/util"
)

type CreateOptions struct {
	Context      string
	NewModel     func() interface{}
	ProcessInput func(context interface{}, model interface{}) error
	handler db.HandlerIface
}

func GetCreateRoute(options *CreateOptions) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, options.Context)
		model := options.NewModel()
		if util.LogWarning(mulekick.Bind(w, r, &model)) {
			return
		}
		if util.LogError(options.ProcessInput(ctx, model)) {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid json",
			})
			return
		}
		if util.LogError(options.handler.Insert(model)) {
			//TODO - should we return this? technically you could probe for an error with this
			mulekick.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "check logs",
			})
			return
		}

		//TODO - send message to event channel

		//Choose your poison, depends where you are in the api currently!!?!!!
		mulekick.WriteJSON(w, http.StatusCreated, model)
		//w.WriteHeader(http.StatusNoContent)
	}
}
