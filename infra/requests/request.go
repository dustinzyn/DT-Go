package requests

import (
	"encoding/json"
	"io/ioutil"
	"strings"

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

// AcceptLanguage .
func (req *Request) AcceptLanguage() (language string) {
	// 支持的语言集
	acceptLangs := map[string]string{
		"zh-CN": "1",
		"zh-TW": "1",
		"en-US": "1",
	}

	// eg. zh-CH, fr;q=0.9, en-US;q=0.8, de;q=0.7, *;q=0.5
	xLanguage := strings.ReplaceAll(req.Worker().Bus().Header.Get("x-language"), " ", "")
	defer func() {
		if r := recover(); r != nil {
			hive.Logger().Errorf("parse x-language: %v, error: %v", xLanguage, r)
			language = utils.GetEnv("LANGUAGE", "zh-CN")
			return
		}
		if language == "" {
			language = utils.GetEnv("LANGUAGE", "zh-CN")
			return
		}
	}()

	langWeights := strings.Split(xLanguage, ",")
	langWeightMap := make(map[string]string)
	for _, langw := range langWeights {
		langWeight := strings.Split(langw, ";")
		switch len(langWeight) {
		case 1:
			// [zh-CH]
			langWeightMap[langWeight[0]] = "1"
		case 2:
			// [fr,q=0.9]
			weight := strings.Split(langWeight[1], "=")[1]
			langWeightMap[langWeight[0]] = weight
		default:
		}
	}
	acceptWeight := ""
	acceptAll := false
	for lang, weight := range langWeightMap {
		// 命中已支持的语言集，并且权重最高的
		if _, ok := acceptLangs[lang]; ok {
			if weight > acceptWeight {
				acceptWeight = weight
				language = lang
			}
		}
		if lang == "*" {
			acceptAll = true
		}
	}
	// 未命中但存在通配符* 采用默认语言
	if language == "" && acceptAll {
		language = utils.GetEnv("LANGUAGE", "zh-CN")
	}
	// TODO 都未命中考虑降级，语言标签去掉区域后再匹配

	return
}

// InternalError 500.
func (req *Request) InternalErrorResponse(err error) hive.Result {
	e := errors.InternalServerError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	return &JSONResponse{Error: e}
}

// BadRequestError 400.
func (req *Request) BadRequestErrorResponse(err error) hive.Result {
	e := errors.BadRequestError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	return &JSONResponse{Error: e}
}

// NoPermissionError 403.
func (req *Request) NoPermissionErrorResponse(err error) hive.Result {
	e := errors.NoPermissionError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	return &JSONResponse{Error: e}
}

// NotFoundError 404.
func (req *Request) NotFoundErrorResponse(err error) hive.Result {
	e := errors.NotFoundError(&errors.ErrorInfo{
		Cause: err.Error(),
	})
	return &JSONResponse{Error: e}
}
