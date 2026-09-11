package route

import (
	"crypto/ecdsa"
	"net/http"
	"strconv"

	"github.com/IceWhaleTech/CasaOS-Common/utils/jwt"
	v1 "github.com/IceWhaleTech/CasaOS-UserService/route/v1"
	"github.com/IceWhaleTech/CasaOS-UserService/service"
	"github.com/labstack/echo/v4"
	echo_middleware "github.com/labstack/echo/v4/middleware"
)

func InitRouter() http.Handler {
	e := echo.New()

	// SECURITY: never trust client-supplied X-Forwarded-For / X-Real-IP
	// headers for remote-address determination (the JWT skipper previously
	// relied on c.RealIP(), which echo derives from those headers).
	e.IPExtractor = echo.ExtractIPDirect()

	e.Use((echo_middleware.CORSWithConfig(echo_middleware.CORSConfig{
		AllowOrigins: []string{
			"http://127.0.0.1:*",
			"http://localhost:*",
			"https://127.0.0.1:*",
			"https://localhost:*",
		},
		AllowMethods:     []string{echo.POST, echo.GET, echo.OPTIONS, echo.PUT, echo.DELETE},
		AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentLength, echo.HeaderXCSRFToken, echo.HeaderContentType, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders, echo.HeaderAccessControlAllowMethods, echo.HeaderConnection, echo.HeaderOrigin, echo.HeaderXRequestedWith},
		ExposeHeaders:    []string{echo.HeaderContentLength, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders},
		MaxAge:           172800,
		AllowCredentials: false,
	})))

	e.Use(echo_middleware.Gzip())

	e.Use(echo_middleware.Logger())

	e.POST("/v1/users/register", v1.PostUserRegister)
	e.POST("/v1/users/login", v1.PostUserLogin)
	e.GET("/v1/users/name", v1.GetUserAllUsername) // all/name
	e.POST("/v1/users/refresh", v1.PostUserRefreshToken)
	// No short-term modifications
	e.GET("/v1/users/image", v1.GetUserImage)

	e.GET("/v1/users/status", v1.GetUserStatus) // init/check

	v1Group := e.Group("/v1")

	v1UsersGroup := v1Group.Group("/users")
	v1UsersGroup.Use(echo_middleware.JWTWithConfig(echo_middleware.JWTConfig{
		// SECURITY: always require a valid JWT — no localhost/loopback skipper.
		// A skipper based on RealIP() is spoofable via X-Forwarded-For and
		// would let any local unprivileged process call these endpoints.
		Skipper: func(c echo.Context) bool {
			return false
		},
		ParseTokenFunc: func(token string, c echo.Context) (interface{}, error) {
			valid, claims, err := jwt.Validate(
				token,
				func() (*ecdsa.PublicKey, error) {
					_, publicKey := service.MyService.User().GetKeyPair()
					return publicKey, nil
				})
			if err != nil || !valid {
				return nil, echo.ErrUnauthorized
			}

			c.Request().Header.Set("user_id", strconv.Itoa(claims.ID))

			return claims, nil
		},
		TokenLookupFuncs: []echo_middleware.ValuesExtractor{
			func(c echo.Context) ([]string, error) {
				// SECURITY: header-only JWT.
				//
				// The query-string token is accepted ONLY on the static asset GET
				// routes that a browser <img> tag loads (which cannot set the
				// Authorization header). It is never accepted on any mutating
				// route. Browsers do not send query params to WebSockets (unlike
				// the container terminal in AppManagement), so none is needed here.
				if len(c.Request().Header.Get(echo.HeaderAuthorization)) > 0 {
					return []string{c.Request().Header.Get(echo.HeaderAuthorization)}, nil
				}
				if isAssetImageRoute(c.Path()) {
					if t := c.QueryParam("token"); t != "" {
						return []string{t}, nil
					}
				}
				return []string{""}, nil
			},
		},
	}))
	{
		v1UsersGroup.Use()
		v1UsersGroup.GET("/current", v1.GetUserInfo)
		v1UsersGroup.PUT("/current", v1.PutUserInfo)
		v1UsersGroup.PUT("/current/password", v1.PutUserPassword)

		v1UsersGroup.GET("/current/custom/:key", v1.GetUserCustomConf)
		v1UsersGroup.POST("/current/custom/:key", v1.PostUserCustomConf)
		v1UsersGroup.DELETE("/current/custom/:key", v1.DeleteUserCustomConf)

		v1UsersGroup.POST("/current/image/:key", v1.PostUserUploadImage)
		v1UsersGroup.PUT("/current/image/:key", v1.PutUserImage)
		// v1UserGroup.POST("/file/image/:key", v1.PostUserFileImage)
		v1UsersGroup.DELETE("/current/image", v1.DeleteUserImage)

		v1UsersGroup.PUT("/avatar", v1.PutUserAvatar)
		v1UsersGroup.GET("/avatar", v1.GetUserAvatar)

		v1UsersGroup.DELETE("/:id", v1.DeleteUser)
		v1UsersGroup.GET("/:username", v1.GetUserInfoByUsername)
		v1UsersGroup.DELETE("", v1.DeleteUserAll)
	}

	return e
}

// isAssetImageRoute reports whether a route serves a static user asset image
// via an <img> tag (which cannot set an Authorization header). The JWT may be
// passed as a short-lived `token` query parameter on these routes ONLY.
func isAssetImageRoute(path string) bool {
	return path == "/v1/users/current/image/:key" || path == "/v1/users/avatar"
}
