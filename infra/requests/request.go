package requests

import (
	"encoding/json"
	"io/ioutil"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
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
func (req *Request) ReadJSON(obj interface{}) (err error) {
	rawData, err := ioutil.ReadAll(req.Worker().IrisContext().Request().Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(rawData, obj)
	if err != nil {
		return
	}
	if err = validate.Struct(obj); err != nil {
		err = errors.BadRequestError(&errors.ErrorInfo{Cause: err.Error()})
		return
	}
	return
}

// ReadQuery .
func (req *Request) ReadQuery(obj interface{}) (err error) {
	if err = req.Worker().IrisContext().ReadQuery(obj); err != nil {
		return
	}
	if err = validate.Struct(obj); err != nil {
		err = errors.BadRequestError(&errors.ErrorInfo{Cause: err.Error()})
		return
	}
	return
}

// ReadForm .
func (req *Request) ReadForm(obj interface{}) (err error) {
	if err = req.Worker().IrisContext().ReadForm(obj); err != nil {
		return
	}
	if err = validate.Struct(obj); err != nil {
		err = errors.BadRequestError(&errors.ErrorInfo{Cause: err.Error()})
		return
	}
	return
}