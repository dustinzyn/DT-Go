package requests

import (
	"encoding/json"
	"io/ioutil"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
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

// ReadQueryDefault .
func (req *Request) ReadQueryDefault(key, defaultValue string) string {
	return req.Worker().IrisContext().URLParamDefault(key, defaultValue)
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

// ReadFormDefault .
func (req *Request) ReadFormDefault(key, defaultValue string) string {
	return req.Worker().IrisContext().FormValueDefault(key, defaultValue)
}

type JSONResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Cause   string      `json:"cause"`
	Detail  interface{} `json:"detail"`
}

// errorResponse .
func (req *Request) errorResponse(err *errors.Error) {
	cStr := utils.IntToStr(err.Code)
	code := utils.StrToInt(cStr[:3])
	ctx := req.Worker().IrisContext()
	ctx.Values().Set("code", code)
	respByte, _ := json.Marshal(err)
	ctx.Values().Set("response", string(respByte))
	ctx.StatusCode(code)
	ctx.JSON(JSONResp{
		Code:    err.Code,
		Message: err.Message,
		Cause:   err.Cause,
		Detail:  err.Detail,
	})
	ctx.StopExecution()
}

// InternalError 500.
func (req *Request) InternalErrorResponse(err error) {
	e := errors.InternalServerError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	req.errorResponse(e)
	return
}

// BadRequestError 400.
func (req *Request) BadRequestErrorResponse(err error) {
	e := errors.BadRequestError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	req.errorResponse(e)
	return
}

// NoPermissionError 403.
func (req *Request) NoPermissionErrorResponse(err error) {
	e := errors.NoPermissionError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	req.errorResponse(e)
	return
}

// NotFoundError 404.
func (req *Request) NotFoundErrorResponse(err error) {
	e := errors.NotFoundError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	req.errorResponse(e)
	return
}
