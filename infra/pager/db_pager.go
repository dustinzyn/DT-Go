package pager

import (
	"fmt"
	"strings"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"gorm.io/gorm"
)

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *Pager {
			return &Pager{}
		})
	})
}

type Builder interface {
	Execute(db *gorm.DB, object interface{}) error
}

type Pager struct {
	pageSize  int
	page      int
	totalPage int
	fields    []string
	items     []string
}

// NewDescPager 多字段降序分页器
func (p *Pager) DescPager(column string, columns ...string) *Pager {
	fieldSort := map[string]string{column: "desc"}
	for _, c := range columns {
		fieldSort[c] = "desc"
	}
	return newDefaultPager(fieldSort)
}

// NewAscPager 多字段升序分页器
func (p *Pager) AscPager(column string, columns ...string) *Pager {
	fieldSort := map[string]string{column: "asc"}
	for _, c := range columns {
		fieldSort[c] = "asc"
	}
	return newDefaultPager(fieldSort)
}

// NewCustomPager 多字段自定义排序分页器
func (p *Pager) CustomPager(fieldSort map[string]string) *Pager {
	return newDefaultPager(fieldSort)
}

// newDefaultPager 默认分页器
func newDefaultPager(fieldSort map[string]string) *Pager {
	fields := make([]string, 0)
	items := make([]string, 0)
	for field, sort := range fieldSort {
		fields = append(fields, field)
		items = append(items, sort)
	}
	return &Pager{
		fields: fields,
		items:  items,
	}
}

// SetPage .
func (p *Pager) SetPage(page, pageSize int) *Pager {
	p.page = page
	p.pageSize = pageSize
	return p
}

// TotalPage .
func (p *Pager) TotalPage() int {
	return p.totalPage
}

// Order 排序
func (p *Pager) Order() interface{} {
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
func (p *Pager) Execute(db *gorm.DB, object interface{}) (err error) {
	pageFind := false
	orderValue := p.Order()
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
