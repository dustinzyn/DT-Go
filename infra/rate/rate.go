package rate

/**
限流组件 基于sentinel-golang实现 https://github.com/alibaba/sentinel-golang

限流配置通过config.ConfigureRateRule来加载

限流熔断机制都是基于资源生效的，不同资源的限流熔断规则互相隔离互不影响。
资源的定义很抽象，用户可以灵活的定义，资源可以是应用、接口、函数、甚至是一段代码。

Created by Dustin.zhu on 2023/08/18.
*/

import (
	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
)

//go:generate mockgen -package mock_infra -source rate.go -destination ./mock/rate_mock.go

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *RateLimitImpl {
			return &RateLimitImpl{}
		})
	})
}

const (
	// 资源类型
	ResTypeCommon     = base.ResTypeCommon
	ResTypeWeb        = base.ResTypeWeb
	ResTypeRPC        = base.ResTypeRPC
	ResTypeAPIGateway = base.ResTypeAPIGateway
	ResTypeDBSQL      = base.ResTypeDBSQL
	ResTypeCache      = base.ResTypeCache
	ResTypeMQ         = base.ResTypeMQ

	// Inbound 入口流量
	Inbound = base.Inbound
	// Outbound 出口流量
	Outbound = base.Outbound
)

type RateLimit interface {
	Entry(resource string, opts ...RateLimitOption) (*base.SentinelEntry, *base.BlockError)
	SetResourceType(base.ResourceType) RateLimit
	SetTrafficType(base.TrafficType) RateLimit
	AddRule(*hive.RateRuleConfiguration) RateLimit
	AddRules([]*hive.RateRuleConfiguration) RateLimit
	GetRules() []*hive.RateRuleConfiguration
}

type RateLimitImpl struct {
	hive.Infra
	resourceType base.ResourceType
	trafficType  base.TrafficType
}

func (rate *RateLimitImpl) BeginRequest(worker hive.Worker) {
	rate.Infra.BeginRequest(worker)
	rate.resourceType = ResTypeCommon
	rate.trafficType = Inbound
}

type RateLimitOption func(*RateLimitImpl)

// WithResourceType sets the resource entry with the given resource type.
func WithResourceType(resourceType base.ResourceType) RateLimitOption {
	return func(opts *RateLimitImpl) {
		opts.resourceType = resourceType
	}
}

// WithTrafficType sets the resource entry with the given traffic type.
func WithTrafficType(entryType base.TrafficType) RateLimitOption {
	return func(opts *RateLimitImpl) {
		opts.trafficType = entryType
	}
}

// Entry is the resource flow limiting entry.
// 通过Entry接口可以把资源包起来，这一步成为“埋点”。通过限流熔断规则来保护资源。每个埋点都有一个资源名称（resource），代表触发了这个资源的调用或访问
// 返回值参数列表的第一个和第二个参数是互斥的，也就是说
// 如果Entry执行pass，那么会返回(*base.SentinelEntry, nil)；
// 如果Entry执行blocked，那么会返回(nil, *base.BlockError)。
func (rate *RateLimitImpl) Entry(resource string, opts ...RateLimitOption) (*base.SentinelEntry, *base.BlockError) {
	for _, opt := range opts {
		opt(rate)
	}
	return sentinel.Entry(resource,
		sentinel.WithResourceType(rate.resourceType),
		sentinel.WithTrafficType(rate.trafficType),
	)
}

// SetResourceType .
func (rate *RateLimitImpl) SetResourceType(typ base.ResourceType) RateLimit {
	rate.resourceType = typ
	return rate
}

// SetTrafficType .
func (rate *RateLimitImpl) SetTrafficType(typ base.TrafficType) RateLimit {
	rate.trafficType = typ
	return rate
}

// AddRule means add a given flow rule to the rule manager
func (rate *RateLimitImpl) AddRule(rule *hive.RateRuleConfiguration) RateLimit {
	flowRule := rate.addRule(rule)
	flowRules := rate.getRules()
	flowRules = append(flowRules, flowRule)
	if _, err := flow.LoadRules(flowRules); err != nil {
		hive.Logger().Infof("Failed to load rules:", err)
	}
	return rate
}

// AddRule means add the given flow rules to the rule manager
func (rate *RateLimitImpl) AddRules(rules []*hive.RateRuleConfiguration) RateLimit {
	flowRules := rate.getRules()
	for _, rule := range rules {
		flowRule := rate.addRule(rule)
		flowRules = append(flowRules, flowRule)
	}
	if _, err := flow.LoadRules(flowRules); err != nil {
		hive.Logger().Infof("Failed to load rules:", err)
	}
	return rate
}

func (rate *RateLimitImpl) addRule(rule *hive.RateRuleConfiguration) *flow.Rule {
	behavior := flow.Reject
	switch rule.ControlBehavior {
	case "Reject":
		behavior = flow.Reject
	case "Throttling":
		behavior = flow.Throttling
	}

	flowRule := flow.Rule{
		Resource:               rule.Resource,
		Threshold:              rule.Threshold,
		TokenCalculateStrategy: flow.Direct,
		ControlBehavior:        behavior,
		StatIntervalInMs:       uint32(rule.StatIntervalInMs),
	}
	if behavior == flow.Throttling {
		flowRule.MaxQueueingTimeMs = uint32(rule.MaxQueueingTimeMs)
	}
	return &flowRule
}

func (rate *RateLimitImpl) getRules() []*flow.Rule {
	sentinelRules := flow.GetRules()
	rules := make([]*flow.Rule, 0)
	for _, srule := range sentinelRules {
		// srule在内存中只会存一份，每次循环只是覆盖值，不能直接使用其指针，需要拷贝
		ssrule := srule
		rules = append(rules, &ssrule)
	}
	return rules
}

func (rate *RateLimitImpl) GetRules() []*hive.RateRuleConfiguration {
	sentinelRules := flow.GetRules()
	rules := make([]*hive.RateRuleConfiguration, 0)
	for _, srule := range sentinelRules {
		rule := hive.RateRuleConfiguration{}
		rule.Resource = srule.Resource
		rule.Threshold = srule.Threshold
		rule.ControlBehavior = srule.ControlBehavior.String()
		rule.TokenCalculateStrategy = srule.TokenCalculateStrategy.String()
		rule.MaxQueueingTimeMs = int(srule.MaxQueueingTimeMs)
		rules = append(rules, &rule)
	}
	return rules
}
