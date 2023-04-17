package infra

import (
	"Hive"
	"encoding/json"
	"io/ioutil"

	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *Request {
			return &Request{}
		})
		initiator.InjectController(func(ctx hive.Context) (com *Request) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

// Request .
type Request struct {
	hive.Infra
}

// BeginRequest .
func (req *Request) BeginRequest(worker hive.Worker) {
	req.Infra.BeginRequest(worker)
}

// ReadJSON .
func (req *Request) ReadJSON(obj interface{}) error {
	rawData, err := ioutil.ReadAll(req.Worker().IrisContext().Request().Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rawData, obj)
	if err != nil {
		return err
	}

	return validate.Struct(obj)
}

// ReadQuery .
func (req *Request) ReadQuery(obj interface{}) error {
	if err := req.Worker().IrisContext().ReadQuery(obj); err != nil {
		return err
	}
	return validate.Struct(obj)
}

// ReadForm .
func (req *Request) ReadForm(obj interface{}) error {
	if err := req.Worker().IrisContext().ReadForm(obj); err != nil {
		return err
	}
	return validate.Struct(obj)
}
