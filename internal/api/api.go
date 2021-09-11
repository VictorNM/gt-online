package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/victornm/gtonline/internal/auth"
	"github.com/victornm/gtonline/internal/gterr"
	"github.com/victornm/gtonline/internal/profile"
)

type API struct {
	Auth    *auth.Service
	Profile *profile.Service
}

func (api *API) register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.RegisterRequest
		if err := api.bind(c, &req); err != nil {
			api.replyErr(c, gterr.New(gterr.InvalidArgument, err.Error(), err))
			return
		}
		res, err := api.Auth.Register(c.Request.Context(), req)
		if err != nil {
			api.replyErr(c, err)
			return
		}

		api.reply(c, 200, res)
	}
}

func (api *API) login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.LoginRequest
		if err := api.bind(c, &req); err != nil {
			api.replyErr(c, gterr.New(gterr.InvalidArgument, err.Error(), err))
			return
		}
		res, err := api.Auth.Login(c.Request.Context(), req)
		if err != nil {
			api.replyErr(c, err)
			return
		}
		api.reply(c, 200, res)
	}
}

func (api *API) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokens := strings.Split(header, " ")
		if len(tokens) != 2 {
			api.abort(c, gterr.New(gterr.Unauthenticated, ""))
			return
		}
		u, err := api.Auth.Authenticate(c.Request.Context(), auth.Token{
			AccessToken: tokens[1],
			TokenType:   tokens[0],
		})
		if err != nil {
			api.abort(c, err)
			return
		}
		c.Set("user", u)
		c.Next()
	}
}

func (api *API) listSchools() gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := api.Profile.ListSchools(c.Request.Context())
		if err != nil {
			api.replyErr(c, err)
			return
		}
		api.reply(c, 200, res)
	}
}

func (api *API) listEmployers() gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := api.Profile.ListEmployers(c.Request.Context())
		if err != nil {
			api.replyErr(c, err)
			return
		}
		api.reply(c, 200, res)
	}
}

func (api *API) getProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := api.userFromContext(c)
		if !ok {
			api.replyErr(c, gterr.New(gterr.Internal, "", fmt.Errorf("can't get User from gin.Context")))
			return
		}

		res, err := api.Profile.GetProfile(c.Request.Context(), profile.GetProfileRequest{
			Email: u.Email,
		})
		if err != nil {
			api.replyErr(c, err)
			return
		}

		api.reply(c, 200, res)
	}
}

func (api *API) updateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(profile.UpdateProfileRequest)
		if err := api.bind(c, &req); err != nil {
			api.replyErr(c, gterr.New(gterr.InvalidArgument, err.Error(), err))
			return
		}

		u, ok := api.userFromContext(c)
		if !ok {
			api.replyErr(c, gterr.New(gterr.Internal, "", fmt.Errorf("can't get User from gin.Context")))
			return
		}
		req.Email = u.Email

		res, err := api.Profile.UpdateProfile(c.Request.Context(), *req)
		if err != nil {
			api.replyErr(c, err)
			return
		}

		api.reply(c, 200, res)
	}
}

func (api *API) userFromContext(c *gin.Context) (*auth.UserAuthDTO, bool) {
	var u *auth.UserAuthDTO
	v, _ := c.Get("user")
	u, ok := v.(*auth.UserAuthDTO)
	return u, ok
}

func (api *API) bind(c *gin.Context, req interface{}) error {
	return c.ShouldBindJSON(req)
}

func (api *API) reply(c *gin.Context, code int, res interface{}) {
	c.JSON(code, res)
}

func (api *API) replyErr(c *gin.Context, err error) {
	_ = c.Error(err)
	e := gterr.Convert(err)
	api.reply(c, httpStatus(e.Code), e)
}

func (api *API) abort(c *gin.Context, err error) {
	api.replyErr(c, err)
	c.Abort()
}

func (api *API) Route(e *gin.Engine) {
	e.POST("/auth/register", api.register())
	e.POST("/auth/login", api.login())

	// Auth endpoints
	e.Use(api.authMiddleware())
	e.GET("/schools", api.listSchools())
	e.GET("/employers", api.listEmployers())
	e.GET("/users/profile", api.getProfile())
	e.PUT("/users/profile", api.updateProfile())
}

func httpStatus(code gterr.ErrorCode) int {
	switch code {
	case gterr.OK:
		return http.StatusOK
	case gterr.Cancelled:
		return 499
	case gterr.Unknown:
		return http.StatusInternalServerError
	case gterr.InvalidArgument:
		return http.StatusBadRequest
	case gterr.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case gterr.NotFound:
		return http.StatusNotFound
	case gterr.AlreadyExists:
		return http.StatusConflict
	case gterr.PermissionDenied:
		return http.StatusForbidden
	case gterr.ResourceExhausted:
		return http.StatusTooManyRequests
	case gterr.FailedPrecondition:
		return http.StatusBadRequest
	case gterr.Aborted:
		return http.StatusConflict
	case gterr.OutOfRange:
		return http.StatusBadRequest
	case gterr.Unimplemented:
		return http.StatusNotImplemented
	case gterr.Internal:
		return http.StatusInternalServerError
	case gterr.Unavailable:
		return http.StatusServiceUnavailable
	case gterr.DataLoss:
		return http.StatusInternalServerError
	case gterr.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
