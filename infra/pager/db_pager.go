package pager

/**
	关系型数据库分页器组件

	1.单字段统一升降序分页器
	2.多字段统一升降序分页器
	3.自定义不同字段的升降序分页器

Created by Dustin.zhu on 2023/05/03.
*/

//go:generate mockgen -package mock_infra -source db_pager.go -destination ./mock/pager_mock.go

import (
	"fmt"
	"strings"

	dt "DT-Go"
	"gorm.io/gorm"
)

func init() {
	dt.Prepare(func(initiator dt.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *PagerImpl {
			return &PagerImpl{}
		})
	})
}

type Pager interface {
	// Execute 执行数据库操作
	Execute(db *gorm.DB, object interface{}) error
	// SetPage 设置分页参数 页数/每页数量
	SetPage(page, pageSize int) Pager
	// TotalPage 总页数
	TotalPage() int
	// DescPager 多字段降序分页器
	DescPager(column string, columns ...string) Pager
	// AscPager 多字段升序分页器
	AscPager(column string, columns ...string) Pager
	// CustomPager 多字段自定义排序分页器
	CustomPager(fieldSort map[string]string) Pager
}

type PagerImpl struct {
	dt.Infra
	pageSize  int
	page      int
	totalPage int
	fields    []string
	items     []string
}

func (p *PagerImpl) BeginRequest(worker dt.Worker) {
	p.Infra.BeginRequest(worker)
}

// DescPager 多字段降序分页器
func (p *PagerImpl) DescPager(column string, columns ...string) Pager {
	fieldSort := map[string]string{column: "desc"}
	for _, c := range columns {
		fieldSort[c] = "desc"
	}
	return newDefaultPager(fieldSort)
}

// AscPager 多字段升序分页器
func (p *PagerImpl) AscPager(column string, columns ...string) Pager {
	fieldSort := map[string]string{column: "asc"}
	for _, c := range columns {
		fieldSort[c] = "asc"
	}
	return newDefaultPager(fieldSort)
}

// CustomPager 多字段自定义排序分页器
func (p *PagerImpl) CustomPager(fieldSort map[string]string) Pager {
	return newDefaultPager(fieldSort)
}

// newDefaultPager 默认分页器
func newDefaultPager(fieldSort map[string]string) Pager {
	fields := make([]string, 0)
	items := make([]string, 0)
	for field, sort := range fieldSort {
		fields = append(fields, field)
		items = append(items, sort)
	}
	return &PagerImpl{
		fields: fields,
		items:  items,
	}
}

// SetPage 设置分页参数 页数/每页数量
func (p *PagerImpl) SetPage(page, pageSize int) Pager {
	p.page = page
	p.pageSize = pageSize
	return p
}

// TotalPage 总页数
func (p *PagerImpl) TotalPage() int {
	return p.totalPage
}

// order 排序
func (p *PagerImpl) order() interface{} {
	if len(p.fields) == 0 {
		return nil
	}
	args := []string{}
	for i := 0; i < len(p.fields); i++ {
		args = append(args, fmt.Sprintf("`%s` %s", p.fields[i], p.items[i]))
	}
	return strings.Join(args, ",")
}

// Execute .
func (p *PagerImpl) Execute(db *gorm.DB, object interface{}) (err error) {
	pageFind := false
	orderValue := p.order()
	if orderValue != nil {
		db = db.Order(orderValue)
	} else {
		db = db.Set("gorm:order_by_primary_key", "DESC")
	}
	if p.page != 0 && p.pageSize != 0 {
		pageFind = true
		db = db.Offset((p.page - 1) * p.pageSize).Limit(p.pageSize)
	}
	resultDB := db.Scan(object)
	if resultDB.Error != nil {
		return resultDB.Error
	}
	if !pageFind {
		return
	}

	var count int64
	err = resultDB.Count(&count).Error
	if err == nil && count != 0 {
		if int(count)%p.pageSize == 0 {
			p.totalPage = int(count) % p.pageSize
		} else {
			p.totalPage = int(count)/p.pageSize + 1
		}
	}
	return
}
