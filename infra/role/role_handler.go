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
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
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
	App        string = "app"
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
type AppAcount struct {
	ID   string `json:"id"`   // 应用账户ID
	Name string `json:"name"` // 应用账户名称
}

type RoleHandler interface {
	// 设置受管控的用户ID
	SetUserID(userID string) RoleHandler
	// 设置允许放行的角色
	SetPermissibleRoles(roles []string) RoleHandler
	// 角色管控
	TrafficOpen() (bool, error)
	// 获取用户信息
	GetUser() User
	// 获取应用账户信息
	GetAppAcount() AppAcount
}

type RoleHandlerImpl struct {
	hive.Infra
	rawUser          User
	rawApp           AppAcount
	permissibleRoles []string // 允许访问的角色
}

// BeginRequest .
func (role *RoleHandlerImpl) BeginRequest(worker hive.Worker) {
	role.permissibleRoles = make([]string, 0)
	role.Infra.BeginRequest(worker)
}

// SetUserID 设置受管控的用户ID
func (role *RoleHandlerImpl) SetUserID(userID string) RoleHandler {
	role.rawUser.ID = userID
	return role
}

// SetPermissibleRoles 设置允许放行的角色
func (role *RoleHandlerImpl) SetPermissibleRoles(roles []string) RoleHandler {
	role.permissibleRoles = append(role.permissibleRoles, roles...)
	return role
}

// TrafficOpen 角色管控 返回true允许访问 返回false禁止访问
func (role *RoleHandlerImpl) TrafficOpen() (bool, error) {
	accountType := role.Infra.Worker().Bus().Get("account_type")
	app := []string{"app"}
	if accountType == "user" {
		user, err := role.getUser()
		if err != nil {
			return false, err
		}
		if len(user.Roles) == 0 {
			return false, err
		}
		if !utils.HasIntersection(role.permissibleRoles, user.Roles) {
			return false, err
		}
	} else {
		if !utils.HasIntersection(role.permissibleRoles, app) {
			return false, nil
		}
		_, err := role.getAppAcount()
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// GetUser 获取用户信息
// 未做角色验证 无法获取用户信息
func (role *RoleHandlerImpl) GetUser() User {
	return role.rawUser
}

// GetAppAcount 获取应用账户信息
// 未做角色验证 无法获取应用账户信息
func (role *RoleHandlerImpl) GetAppAcount() AppAcount {
	return role.rawApp
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

func getAppUrl(appId string) string {
	cg := hive.NewConfiguration()

	url := url.URL{
		Scheme: cg.DS.UserMgntProtocol,
		Host:   fmt.Sprintf("%v:%v", cg.DS.UserMgntHost, cg.DS.UserMgntPort),
	}
	url.Path = fmt.Sprintf("/api/user-management/v1/apps/%s", appId)
	return url.String()
}

// GetUser 获取用户信息
func (role *RoleHandlerImpl) getUser() (user User, err error) {
	ownerEndpoint := getOwnersEndpoint(role.rawUser.ID)
	users := make([]User, 1)
	users[0] = User{}
	resp := hivehttp.NewHTTPRequest(ownerEndpoint).Get().ToJSON(&users)
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New(role.Worker().Bus().Get("language"), errors.ResourceNotFoundErr, "User not found", nil)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get user error: status code = %v, response err = %v", resp.StatusCode, resp.Error)
		return
	}
	user = users[0]
	role.rawUser = user
	return
}

// getAppAcount 获取应用账户信息
func (role *RoleHandlerImpl) getAppAcount() (appAcount AppAcount, err error) {
	appUrl := getAppUrl(role.rawUser.ID)
	resp := hivehttp.NewHTTPRequest(appUrl).Get().ToJSON(&appAcount)
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New(role.Worker().Bus().Get("language"), errors.ResourceNotFoundErr, "App not found", nil)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get app error: status code = %v, response err = %v", resp.StatusCode, resp.Error)
		return
	}
	role.rawApp = appAcount
	return
}
