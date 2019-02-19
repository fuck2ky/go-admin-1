package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/hequan2017/go-admin/docs"
	"github.com/hequan2017/go-admin/middleware/inject"
	"github.com/hequan2017/go-admin/middleware/jwt"
	"github.com/hequan2017/go-admin/middleware/permission"
	"github.com/hequan2017/go-admin/pkg/setting"
	"github.com/hequan2017/go-admin/routers/api"
	"github.com/hequan2017/go-admin/routers/api/v1"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"net/http"
	"strings"
)

func InitRouter() *gin.Engine {
	obj := inject.Init()

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(Cors())

	r.Use(gin.Recovery())
	gin.SetMode(setting.ServerSetting.RunMode)

	err := loadCasbinPolicyData(obj)
	if err != nil {
		panic("加载casbin策略数据发生错误: " + err.Error())
	}

	r.POST("/auth", api.Auth)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiv1 := r.Group("/api/v1")
	apiv1.Use(jwt.JWT())
	apiv1.Use(permission.CasbinMiddleware(obj.Enforcer))
	{

		apiv1.GET("/menus", v1.GetMenus)
		apiv1.GET("/menus/:id", v1.GetMenu)
		apiv1.POST("/menus", v1.AddMenu)
		apiv1.PUT("/menus/:id", v1.EditMenu)
		apiv1.DELETE("/menus/:id", v1.DeleteMenu)

		apiv1.GET("/roles", v1.GetRoles)
		apiv1.GET("/roles/:id", v1.GetRole)
		apiv1.POST("/roles", v1.AddRole)
		apiv1.PUT("/roles/:id", v1.EditRole)
		apiv1.DELETE("/roles/:id", v1.DeleteRole)

		apiv1.GET("/users", api.GetUsers)
		apiv1.GET("/users/:id", api.GetUser)
		apiv1.POST("/users", api.AddUser)
		apiv1.PUT("/users/:id", api.EditUser)
		apiv1.DELETE("/users/:id", api.DeleteUser)
	}

	return r
}

// 加载casbin策略数据，包括角色权限数据、用户角色数据
func loadCasbinPolicyData(obj *inject.Object) error {
	c := obj.Common

	err := c.RoleAPI.LoadAllPolicy()
	if err != nil {
		return err
	}
	err = c.UserAPI.LoadAllPolicy()
	if err != nil {
		return err
	}
	return nil
}

// 跨域
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method      //请求方法
		origin := c.Request.Header.Get("Origin")        //请求头部
		var headerKeys []string                             // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")        // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")      //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")      // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")        // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")       //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")       // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next()        //  处理请求
	}
}