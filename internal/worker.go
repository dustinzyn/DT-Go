package internal

import (
	stdContext "context"
	"math/rand"
	"net/http"
	"reflect"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/memstore"
)

const (
	// WorkerKey specify where is the Worker should be located in Context
	WorkerKey = "STORE-WORKER-KEY"
)

// Worker describe a global context which use to share the internal component
// (infrastructure, logger and so on) with middleware, controller, domain service and etc.
type Worker interface {

	// IrisContext point to current iris.Context instance
	IrisContext() iris.Context

	// Logger return current logger
	Logger() Logger

	// Store return an address of current memstore.Store
	// memstore.Store is a collection of key/value entries.
	// usually use to store metadata produced by service runtime.
	Store() *memstore.Store

	// Bus return an address of current bus
	Bus() *Bus

	// Context return current context
	Context() stdContext.Context

	// WithContext set current context instead Context()
	WithContext(stdContext.Context)

	// StartTime() return a time since this this Worker be created
	StartTime() time.Time

	// DeferRecycle marks the resource wont't be recycle immediately after the request has ended
	//
	// here is a simple explain about this
	//
	// When an Http request is incoming, the program will probably serve a bunch of business
	// logic services, DB connetion, transaction, redis caches, and so on. Once those producer
	// has done, the system should write response and release those resource immediately.
	// In other words, the system should do some clean up procedures for this request.
	// You might thought it is a matter of course. But in special cases, such as goroutine
	// without synchronizing-signal. When all business procedures has completed on businesses
	// goroutine, and prepare to respond. GC system may be run before the http handler goroutine
	// to respond the client. Once this apportunity was met, the client will got an "Internal Server Error"
	// or other wrong result, because resource has been recycled by GC to repond to client.
	DeferRecycle()

	// IsDeferRecycle() indicates system need to wait while for recycle resource
	IsDeferRecycle() bool

	// Rand return a rand.Rand act a random number seeder.
	Rand() *rand.Rand

	// IsPrivate returns true means the application is private, otherwise is public
	IsPrivate() bool
}

func newWorkerHandle(private bool) context.Handler {
	return func(ctx *context.Context) {
		work := newWorker(ctx, private)
		ctx.Values().Set(WorkerKey, work)
		ctx.Next()

		if work.IsDeferRecycle() {
			return
		}
		work.logger = nil
		work.ctx = nil
		ctx.Values().Reset()
	}
}

func newWorker(ctx iris.Context, private bool) *worker {
	work := new(worker)
	work.ctx = ctx
	work.private = private
	work.services = make([]interface{}, 0)
	work.coms = make([]interface{}, 0)
	head := ctx.Request().Header
	if head == nil {
		head = make(http.Header)
	}
	work.bus = newBus(head)
	work.stdCtx = ctx.Request().Context()
	work.time = time.Now()
	work.deferRecycle = false
	HandlerBusMiddleware(work)
	return work
}

type worker struct {
	ctx          iris.Context
	private      bool
	services     []interface{}
	coms         []interface{}
	logger       Logger
	bus          *Bus
	stdCtx       stdContext.Context
	time         time.Time
	values       memstore.Store
	deferRecycle bool
	randInstance *rand.Rand
}

func (w *worker) IrisContext() iris.Context {
	return w.ctx
}

func (w *worker) Logger() Logger {
	if w.logger == nil {
		l := w.values.Get("logger_impl")
		if l == nil {
			if w.IsPrivate() {
				w.logger = privateApp.Logger()
			} else {
				w.logger = publicApp.Logger()
			}
		} else {
			w.logger = l.(Logger)
		}
	}
	return w.logger
}

func (w *worker) Context() stdContext.Context {
	return w.stdCtx
}

func (w *worker) WithContext(ctx stdContext.Context) {
	w.stdCtx = ctx
}

func (w *worker) StartTime() time.Time {
	return w.time
}

func (w *worker) Store() *memstore.Store {
	return &w.values
}

func (w *worker) Bus() *Bus {
	return w.bus
}

func (w *worker) DeferRecycle() {
	w.deferRecycle = true
}

func (w *worker) IsDeferRecycle() bool {
	return w.deferRecycle
}

func (w *worker) Rand() *rand.Rand {
	if w.randInstance == nil {
		w.randInstance = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return w.randInstance
}

func (w *worker) SetLogger(l Logger) {
	w.logger = l
}

func (w *worker) IsPrivate() bool {
	return w.private
}

var workerType = reflect.TypeOf(&worker{})
