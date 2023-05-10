package hive

import (
	"io/ioutil"
	"os"

	redis "github.com/go-redis/redis/v8"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/hero"
	"github.com/kataras/iris/v12/mvc"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/internal"
	"gopkg.in/yaml.v3"
)

var privateApp *internal.Application
var publicApp *internal.Application

func init() {
	publicApp = internal.NewPublicApplication()
	privateApp = internal.NewPrivateApplication()
}

type (
	// Worker .
	Worker = internal.Worker

	// Initiator .
	Initiator = internal.Initiator

	//Starter is the startup interface.
	Starter = internal.Starter

	// Infra .
	Infra = internal.Infra

	//SingleBoot .
	SingleBoot = internal.SingleBoot

	//Entity is the entity's father interface.
	Entity = internal.Entity

	// BusHandler is the bus message middleware type.
	BusHandler = internal.BusHandler

	//Repository .
	Repository = internal.Repository

	//Result is the controller return type.
	Result = hero.Result

	//Context is the context type.
	Context = iris.Context

	// Bus is the bus message type.
	Bus = internal.Bus

	// Configuration is the configuration type of the app.
	Configuration = iris.Configuration

	// DomainEvent .
	DomainEvent = internal.DomainEvent

	// LogRow is the log per line callback.
	LogRow = golog.Log

	// BeforeActivation is Is the start-up pre-processing of the action.
	BeforeActivation = mvc.BeforeActivation

	// LogFields is the column type of the log.
	LogFields = golog.Fields
)

// NewPublicApplication returns Application interface type
func NewPublicApplication() Application {
	return publicApp
}

// NewPrivateApplication returns Application interface type
func NewPrivateApplication() Application {
	return privateApp
}

// Application
type Application interface {
	InstallDB(f func() (db interface{}))
	InstallDBTable(f func() (tables map[string]interface{}))
	InstallRedis(f func() (client redis.Cmdable))
	InstallOther(f func() interface{})
	InstallMiddleware(handler iris.Handler)
	InstallParty(relativePath string)
	NewRunner(addr string, configurators ...host.Configurator) iris.Runner
	NewH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
	NewAutoTLSRunner(addr, domain, email string, configurators ...host.Configurator) iris.Runner
	NewTLSRunner(addr, certFile, keyFile string, configurators ...host.Configurator) iris.Runner
	Iris() *iris.Application
	Logger() *golog.Logger
	Run(serve iris.Runner, c iris.Configuration)
	Start(f func(starter Starter))
	InstallBusMiddleware(handler ...BusHandler)
	InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
	CallService(fun interface{}, worker ...Worker)
}

func Prepare(f func(Initiator)) {
	internal.Prepare(f)
}

func Logger() *golog.Logger {
	if publicApp != nil {
		return publicApp.Logger()
	} else {
		return privateApp.Logger()
	}
}

// Configure
type Configurer interface {
	Configure(obj interface{}, file string, metadata ...interface{}) error
}

var configurer Configurer

// SetConfigurer
func SetConfigurer(confer Configurer) {
	configurer = confer
}

var ProfileENV = "CONFIG_PATH"

// Configure
func Configure(obj interface{}, file string, metadata ...interface{}) (err error) {
	if configurer != nil {
		return configurer.Configure(obj, file, metadata...)
	}
	path := os.Getenv(ProfileENV)
	if path == "" {
		path = "./conf"
		if _, err := os.Stat(path); err != nil {
			path = "./server/conf"
			if _, err := os.Stat(path); err != nil {
				path = ""
			}
		}
	}
	ioStream, err := ioutil.ReadFile(path + "/" + file)
	if err != nil {
		Logger().Errorf("Configure readfile error: %v", err)
	}
	err = yaml.Unmarshal(ioStream, obj)
	if err != nil {
		Logger().Errorf("Configure decode error: %s", err.Error())
	} else {
		Logger().Infof("Configure decode: %s", path+"/"+file)
	}
	return
}

func ToWorker(ctx Context) Worker {
	if result, ok := ctx.Values().Get(internal.WorkerKey).(Worker); ok {
		return result
	}
	return nil
}

// DefaultConfiguration the default profile.
func DefaultConfiguration() iris.Configuration {
	return iris.DefaultConfiguration()
}
