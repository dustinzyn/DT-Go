package requests

import (
	"encoding/json"
	"io/ioutil"
	"mime/multipart"

	dhive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
	"github.com/kataras/iris/v12"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	dhive.Prepare(func(initiator dhive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *RequestImpl {
			return &RequestImpl{}
		})
		initiator.InjectController(func(ctx dhive.Context) (com *RequestImpl) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

// Request .
type Request interface {
	ReadJSON(obj interface{}) (err error)
	ReadQuery(obj interface{}) (err error)
	ReadQueryDefault(key, defaultValue string) string
	ReadForm(obj interface{}) (err error)
	ReadFormDefault(key, defaultValue string) string
	AcceptLanguage() (language string)
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	FormFiles(key string, before ...func(dhive.Context, *multipart.FileHeader) bool) (files []multipart.File, headers []*multipart.FileHeader, err error)
}

// RequestImpl .
type RequestImpl struct {
	dhive.Infra
}

// BeginRequest .
func (req *RequestImpl) BeginRequest(worker dhive.Worker) {
	req.Infra.BeginRequest(worker)
}

// ReadJSON .
func (req *RequestImpl) ReadJSON(obj interface{}) (err error) {
	rawData, err := ioutil.ReadAll(req.Worker().IrisContext().Request().Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(rawData, obj)
	if err != nil {
		err = errors.New(req.AcceptLanguage(), errors.BadRequestErr, err.Error(), nil)
		return
	}
	if err = validate.Struct(obj); err != nil {
		err = errors.New(req.AcceptLanguage(), errors.BadRequestErr, err.Error(), nil)
		return
	}
	return
}

// ReadQuery .
func (req *RequestImpl) ReadQuery(obj interface{}) (err error) {
	if err = req.Worker().IrisContext().ReadQuery(obj); err != nil {
		return
	}
	if err = validate.Struct(obj); err != nil {
		err = errors.New(req.AcceptLanguage(), errors.BadRequestErr, err.Error(), nil)
		return
	}
	return
}

// ReadQueryDefault .
func (req *RequestImpl) ReadQueryDefault(key, defaultValue string) string {
	return req.Worker().IrisContext().URLParamDefault(key, defaultValue)
}

// ReadForm .
func (req *RequestImpl) ReadForm(obj interface{}) (err error) {
	if err = req.Worker().IrisContext().ReadForm(obj); err != nil && !iris.IsErrPath(err) {
		err = errors.New(req.AcceptLanguage(), errors.BadRequestErr, err.Error(), nil)
		return
	}
	if err = validate.Struct(obj); err != nil {
		err = errors.New(req.AcceptLanguage(), errors.BadRequestErr, err.Error(), nil)
		return
	}
	return
}

// ReadFormDefault .
func (req *RequestImpl) ReadFormDefault(key, defaultValue string) string {
	return req.Worker().IrisContext().FormValueDefault(key, defaultValue)
}

// AcceptLanguage .
func (req *RequestImpl) AcceptLanguage() (language string) {
	language = utils.ParseXLanguage(req.Worker().Bus().Header.Get("x-language"))
	// 注入消息总线
	req.Worker().Bus().Add("language", language)
	return
}

// FormFile .
func (req *RequestImpl) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return req.Worker().IrisContext().FormFile(key)
}

// FormFile .
func (req *RequestImpl) FormFiles(key string, before ...func(dhive.Context, *multipart.FileHeader) bool) (files []multipart.File, headers []*multipart.FileHeader, err error) {
	return req.Worker().IrisContext().FormFiles(key, before...)
}
