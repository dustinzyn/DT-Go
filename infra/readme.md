## 如何自定义基础设施组件
* 单例组件入口是 Booting, 生命周期为常驻
* 多例组件入口是 BeginRequest，生命周期为一个请求会话
* 框架已提供的组件目录 devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DT-Go/infra
* 用户自定义的组件目录 [project]/infra/[custom]
* 组件可以独立使用组件的配置文件, 配置文件放在 [project]/conf/infra/[*.yaml]

## 单例组件
``` golang
func init() {
	dt.Prepare(func(initiator dt.Initiator) {
		/*
			绑定组件
			single 是否单例
			com :如果是单例com是组件指针， 如果是多例 com是创建组件的函数
			BindInfra(single bool, com interface{})
		*/
		initiator.BindInfra(true, initiator.IsPrivate(), &Single{})

		/*
			该组件注入到控制器, 默认仅注入到service和repository
			如果不调用 initiator.InjectController, 控制器无法使用。
		*/
		initiator.InjectController(func(ctx dt.Context) (com *Single) {
			initiator.FetchInfra(ctx, &com)
			return
		})
	})
}

type Single struct {
	life int
}

// Booting 单例组件入口, 启动时调用一次。
func (c *Single) Booting(boot dt.SingleBoot) {
	dt.Logger().Info("Single.Booting")
	c.life = rand.Intn(100)
}

func (mu *Single) GetLife() int {
	//所有请求的访问 都是一样的life
	return mu.life
}
```

## 多例组件
实现一个读取json数据的组件，并且做数据验证。
``` golang
//使用展示
type GoodsController struct {
	JSONRequest *infra.JSONRequest
}
// GetBy handles the PUT: /goods/stock route 增加商品库存.
func (goods *GoodsController) PutStock() dt.Result {
	var request struct {
		GoodsID int `json:"goodsId" validate:"required"` //商品id
		Num     int `validate:"min=1,max=15"`            //只能增加的范围1-15，其他报错
	}

	//使用自定义的json组件读取请求数据, 并且处理数据验证。
	if e := goods.JSONRequest.ReadJSON(&request); e != nil {
		return &infra.JSONResponse{Err: e}
	}
}
```
``` golang
//组件的实现
func init() {
	validate = validator.New()
	dt.Prepare(func(initiator dt.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *JSONRequest {
			//绑定1个New多例组件的回调函数，多例目的是为每个请求独立服务。
			return &JSONRequest{}
		})
		initiator.InjectController(func(ctx dt.Context) (com *JSONRequest) {
			//从Infra池里取出注入到控制器。
			initiator.FetchInfra(ctx, &com)
			return
		})
	})
}

// JSONRequest .
type JSONRequest struct {
	dt.Infra	//多例需继承dt.Infra
}

// BeginRequest 每一个请求只会触发一次
func (req *JSONRequest) BeginRequest(worker dt.Worker) {
	// 调用基类初始化请求运行时
	req.Infra.BeginRequest(worker)
}

// ReadJSON .
func (req *JSONRequest) ReadJSON(obj interface{}) error {
	//从上下文读取io数据
	rawData, err := ioutil.ReadAll(req.Worker().Ctx().Request().Body)
	if err != nil {
		return err
	}

	/*
		使用第三方 validate 做数据验证
	*/
	err = json.Unmarshal(rawData, obj)
	if err != nil {
		return err
	}

	return validate.Struct(obj)
}
```
