package role

/**
角色管控组件
管控不同角色的访问权限

Created by Dustin.zhu on 2023/05/03.
*/

//go:generate mockgen -package mock_infra -source role_handler.go -destination ./mock/role_mock.go

import (
	"fmt"
	"net/http"
	"net/url"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/hivehttp"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
)

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *RoleHandlerImpl {
			return &RoleHandlerImpl{}
		})
		/*
			注入到控制器, 默认仅注入到service和repository
			如果不调用 initiator.InjectController, 控制器无法使用。
		*/
		initiator.InjectController(func(ctx hive.Context) (com *RoleHandlerImpl) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

// 角色常量的定义
const (
	SuperAdmin string = "super_admin"
	SysAdmin   string = "sys_admin"
	AuditAdmin string = "audit_admin"
	SecAdmin   string = "sec_admin"
	OrgManager string = "org_manager"
	OrgAudit   string = "org_audit"
	NormalUser string = "normal_user"
)

type User struct {
	ID       string   `json:"id"`        // 用户ID
	Roles    []string `json:"roles"`     // 角色
	Name     string   `json:"name"`      // 名称
	CsfLevel float64  `json:"csf_level"` // 密级
	Enabled  bool     `json:"enabled"`   // 是否启用
	Frozen   bool     `json:"frozen"`    // 冻结状态
	Email    string   `json:"email"`     // 邮箱
	AuthType string   `json:"auth_type"` // 认证类型
}

type RoleHandler interface {
	// 设置受管控的用户ID
	SetUserID(userID string) RoleHandler
	// 设置允许放行的角色
	SetPermissibleRoles(roles []string) RoleHandler
	// 角色管控
	TrafficOpen() (bool, error)
}

type RoleHandlerImpl struct {
	hive.Infra
	UserID           string   // 用户ID
	PermissibleRoles []string // 允许访问的角色
}

// BeginRequest .
func (role *RoleHandlerImpl) BeginRequest(worker hive.Worker) {
	role.PermissibleRoles = make([]string, 0)
	role.Infra.BeginRequest(worker)
}

// SetUserID 设置受管控的用户ID
func (role *RoleHandlerImpl) SetUserID(userID string) RoleHandler {
	role.UserID = userID
	return role
}

// SetPermissibleRoles 设置允许放行的角色
func (role *RoleHandlerImpl) SetPermissibleRoles(roles []string) RoleHandler {
	role.PermissibleRoles = append(role.PermissibleRoles, roles...)
	return role
}

// TrafficOpen 角色管控 返回true允许访问 返回false禁止访问
func (role *RoleHandlerImpl) TrafficOpen() (bool, error) {
	user, err := role.userInfo(role.UserID)
	if err != nil {
		return false, err
	}
	if !utils.HasIntersection(role.PermissibleRoles, user.Roles) {
		return false, err
	}
	return true, nil
}

func getOwnersEndpoint(userId string) string {
	cg := hive.NewConfiguration()

	url := url.URL{
		Scheme: cg.DS.UserMgntProtocol,
		Host:   fmt.Sprintf("%v:%v", cg.DS.UserMgntHost, cg.DS.UserMgntPort),
	}
	url.Path = fmt.Sprintf("/api/user-management/v1/users/%v/roles,csf_level,name,enabled,frozen,auth_type", userId)
	return url.String()
}

// userInfo 获取用户信息
func (role *RoleHandlerImpl) userInfo(userID string) (user User, err error) {
	ownerEndpoint := getOwnersEndpoint(userID)
	users := make([]User, 1)
	users[0] = User{}
	resp := hivehttp.NewHTTPRequest(ownerEndpoint).Get().ToJSON(&users)
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get user error: status code = %v, response err = %v", resp.StatusCode, resp.Error)
		return
	}
	user = users[0]
	return
}
