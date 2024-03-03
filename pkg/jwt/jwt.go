package jwt

import (
	"strings"
	"time"

	"github.com/achwanyusuf/carrent-lib/pkg/errormsg"
	"github.com/achwanyusuf/carrent-lib/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt"
)

type authHeader struct {
	IDToken string `header:"Authorization"`
}

type transactionInfo struct {
	RequestURI    string    `json:"request_uri"`
	RequestMethod string    `json:"request_method"`
	RequestID     string    `json:"request_id"`
	Timestamp     time.Time `json:"timestamp"`
	ErrorCode     int64     `json:"error_code,omitempty"`
	Cause         string    `json:"cause,omitempty"`
}

type auth struct {
	AccessToken string     `json:"access_token,omitempty"`
	TokenType   string     `json:"token_type,omitempty"`
	Exp         *time.Time `json:"exp,omitempty"`
	Scope       string     `json:"scope,omitempty"`
}

type response struct {
	TransactionInfo transactionInfo `json:"transaction_info"`
	Code            int64           `json:"status_code"`
	Message         string          `json:"message,omitempty"`
	Translation     *translation    `json:"translation,omitempty"`
}

type translation struct {
	EN string `json:"en"`
}

type loginResponse struct {
	response
	auth
}

func (l *loginResponse) Transform(ctx *gin.Context, log logger.Logger, code int, err error) int {
	l.response = response{
		TransactionInfo: transactionInfo{
			RequestURI:    ctx.Request.RequestURI,
			RequestMethod: ctx.Request.Method,
			RequestID:     ctx.GetHeader("x-request-id"),
			Timestamp:     time.Now(),
		},
		Code: int64(code),
	}
	if err != nil {
		getErrMsg := errormsg.GetErrorData(err)
		l.response.TransactionInfo.ErrorCode = getErrMsg.Code
		log.Error(ctx, errormsg.WriteErr(err))
		l.response.Code = getErrMsg.WrappedMessage.StatusCode
		l.response.Message = getErrMsg.WrappedMessage.Message
		translation := translation(getErrMsg.WrappedMessage.Translation)
		l.response.Translation = &translation
	}

	return int(l.response.Code)
}

func JWT(log logger.Logger, signingKey []byte) gin.HandlerFunc {
	resp := loginResponse{}
	return func(ctx *gin.Context) {
		h := authHeader{}
		if err := ctx.ShouldBindHeader(&h); err != nil {
			if _, ok := err.(validator.ValidationErrors); ok {
				err := errormsg.WrapErr(errormsg.Error400, err, "invalid request parameters")
				resp.Transform(ctx, log, int(errormsg.Error400.StatusCode), err)

				ctx.JSON(int(errormsg.Error400.StatusCode), resp)
				ctx.Abort()
				return
			}
			err := errormsg.WrapErr(errormsg.Error400, err, "invalid request parameters")
			resp.Transform(ctx, log, int(errormsg.Error400.StatusCode), err)

			ctx.JSON(int(errormsg.Error400.StatusCode), resp)
			ctx.Abort()
			return
		}

		tokenStr := strings.Split(h.IDToken, "Bearer ")

		if len(tokenStr) < 2 {
			err := errormsg.WrapErr(errormsg.Error401, nil, "authorization header should be with prefix Bearer")
			resp.Transform(ctx, log, int(errormsg.Error401.StatusCode), err)

			ctx.JSON(int(errormsg.Error401.StatusCode), resp)
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr[1], func(token *jwt.Token) (interface{}, error) {
			// check token signing method etc
			return signingKey, nil
		})
		if err != nil {
			err := errormsg.WrapErr(errormsg.Error401, err, "invalid token")
			resp.Transform(ctx, log, int(errormsg.Error401.StatusCode), err)

			ctx.JSON(int(errormsg.Error401.StatusCode), resp)
			ctx.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx.Set("id", int64(claims["id"].(float64)))
			ctx.Set("username", claims["username"])
			ctx.Set("scope", claims["scope"])
		} else {
			err := errormsg.WrapErr(errormsg.Error401, nil, "invalid claims")
			resp.Transform(ctx, log, int(errormsg.Error401.StatusCode), err)

			ctx.JSON(int(errormsg.Error401.StatusCode), resp)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
