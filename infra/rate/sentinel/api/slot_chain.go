// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/base"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/circuitbreaker"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/flow"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/hotspot"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/isolation"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/log"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/stat"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/system"
)

var globalSlotChain = BuildDefaultSlotChain()

func GlobalSlotChain() *base.SlotChain {
	return globalSlotChain
}

func BuildDefaultSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlot(stat.DefaultResourceNodePrepareSlot)

	sc.AddRuleCheckSlot(system.DefaultAdaptiveSlot)
	sc.AddRuleCheckSlot(flow.DefaultSlot)
	sc.AddRuleCheckSlot(isolation.DefaultSlot)
	sc.AddRuleCheckSlot(hotspot.DefaultSlot)
	sc.AddRuleCheckSlot(circuitbreaker.DefaultSlot)

	sc.AddStatSlot(stat.DefaultSlot)
	sc.AddStatSlot(log.DefaultSlot)
	sc.AddStatSlot(flow.DefaultStandaloneStatSlot)
	sc.AddStatSlot(hotspot.DefaultConcurrencyStatSlot)
	sc.AddStatSlot(circuitbreaker.DefaultMetricStatSlot)
	return sc
}