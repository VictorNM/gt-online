package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/victornm/gtonline/internal/auth"
	"github.com/victornm/gtonline/internal/gterr"
)

type API struct {
	Auth *auth.Service
}

func (api *API) register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			api.reply(c, http.StatusBadRequest, gterr.New(gterr.InvalidArgument, err.Error(), err))
			return
		}
		res, err := api.Auth.Register(c.Request.Context(), req)
		if err != nil {
			e := gterr.Convert(err)
			api.reply(c, httpStatus(e.Code), e)
			return
		}

		api.reply(c, 200, res)
	}
}

func (api *API) reply(c *gin.Context, code int, obj interface{}) {
	if err, ok := obj.(*gterr.Error); ok && err.Detail != nil {
		_ = c.Error(err)
	}
	c.JSON(code, obj)
}

func (api *API) Route(e *gin.Engine) {
	e.POST("/auth/register", api.register())
}

func httpStatus(code gterr.ErrorCode) int {
	switch code {
	case gterr.OK:
		return http.StatusOK
	case gterr.InvalidArgument:
		return http.StatusBadRequest
	case gterr.AlreadyExists:
		return http.StatusConflict
	case gterr.Unknown:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
