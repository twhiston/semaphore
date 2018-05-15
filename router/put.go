package router

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"net/http"
)

type PatchRequestOptions struct {
	Context      string
	NewModel     func() interface{}
	ProcessInput func(context interface{}, model interface{}) error
}

func GetPutRoute(options *PatchRequestOptions) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, options.Context)
		model := options.NewModel()
		if err := mulekick.Bind(w, r, &model); err != nil {
			util.LogErrorWithFields(err, logrus.Fields{"context": ctx, "data": model})
		}
		if util.LogError(options.ProcessInput(ctx, model)) {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid json",
			})
			return
		}
		rows, err := db.Connection.GetHandler().Update(model)
		if util.LogError(err) {
			//TODO - should we return this? technically you could probe for an error with this
			mulekick.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "check logs",
			})
			return
		}
		if rows == 0 {
			util.LogWarningWithFields(errors.New("could not update object"), logrus.Fields{"context": ctx, "model": model})
			return
		}
		if rows > 1 {
			util.LogPanicWithFields(errors.New("patch action resulted in database corruption"), logrus.Fields{"rows": rows, "context": ctx, "model": model})
			//dead
		}

		// TODO - write event to channel

		w.WriteHeader(http.StatusNoContent)
	}

}
