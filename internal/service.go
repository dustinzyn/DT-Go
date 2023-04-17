package internal

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/kataras/iris/v12/context"
)

func callService(private bool, fun interface{}, worker ...Worker) {
	var app *Application
	if private {
		app = privateApp
	} else {
		app = publicApp
	}
	if len(worker) == 0 {
		worker = make([]Worker, 1)
		ctx := context.NewContext(app.IrisApp)
		ctx.BeginRequest(nil, new(http.Request))
		rt := newWorker(ctx, private)
		ctx.Values().Set(WorkerKey, rt)
		worker[0] = rt
	}
	serviceObj, err := parseCallServiceFunc(fun)
	if err != nil {
		panic(fmt.Sprintf("CallService, %v : %s", fun, err.Error()))
	}
	newService := app.pool.create(worker[0], serviceObj)
	reflect.ValueOf(fun).Call([]reflect.Value{reflect.ValueOf(newService.(serviceElement).serviceObject)})

	if worker[0].IsDeferRecycle() {
		return
	}
	app.pool.free(newService)
}
