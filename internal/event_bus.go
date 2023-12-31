package internal

import "reflect"

type EventBus struct {
	private        bool
	eventsPath     map[string]string
	eventsAddr     map[string]string
	controllers    []interface{}
	eventsInfraCom map[string]reflect.Type
}

func NewEventBus(private bool) *EventBus {
	return &EventBus{
		private:        private,
		eventsPath:     make(map[string]string),
		eventsAddr:     make(map[string]string),
		eventsInfraCom: make(map[string]reflect.Type),
	}
}

func (msgBus *EventBus) addController(controller interface{}) {
	msgBus.controllers = append(msgBus.controllers, controller)
}

func (msgBus *EventBus) app() *Application {
	if msgBus.private {
		return privateApp
	} else {
		return publicApp
	}
}

func (msgBus *EventBus) addEvent(objectMethod, eventName string, infraCom ...interface{}) {
	if _, ok := msgBus.eventsAddr[eventName]; ok {
		msgBus.app().Logger().Fatalf("ListenEvent: Event already bound :%v", eventName)
	}
	msgBus.eventsAddr[eventName] = objectMethod
	if len(infraCom) > 0 {
		infraComType := reflect.TypeOf(infraCom[0])
		msgBus.eventsInfraCom[eventName] = infraComType
	}
}

// EventsPath .
func (msgBus *EventBus) EventsPath(infra interface{}) (msgs map[string]string) {
	infraComType := reflect.TypeOf(infra)
	msgs = make(map[string]string)
	for k, v := range msgBus.eventsPath {
		ty, ok := msgBus.eventsInfraCom[k]
		if ok && ty != infraComType {
			continue
		}
		msgs[k] = v
	}
	return
}

func (msgBus *EventBus) building() {
	eventsRoute := make(map[string]string)
	for _, controller := range msgBus.controllers {
		v := reflect.ValueOf(controller)
		t := reflect.TypeOf(controller)
		for index := 0; index < v.NumMethod(); index++ {
			method := t.Method(index)
			eventName := msgBus.match(t.Elem().Name() + "." + method.Name)
			if eventName == "" {
				continue
			}
			eventsRoute[eventName] = t.Elem().String() + "." + method.Name
		}
	}
	for _, r := range msgBus.app().IrisApp.GetRoutes() {
		for eventName, handlersName := range eventsRoute {
			if r.MainHandlerName != handlersName {
				continue
			}
			if r.Method != "POST" {
				msgBus.app().Logger().Fatalf("ListenEvent: Event routing must be 'post', MainHandlerName:%v", r.MainHandlerName)
			}
			msgBus.eventsPath[eventName] = r.Path
		}
	}
}

func (msgBus *EventBus) match(objectMethod string) (eventName string) {
	for name, addrs := range msgBus.eventsAddr {
		if addrs == objectMethod {
			eventName = name
			return
		}
	}
	return
}
