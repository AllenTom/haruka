package middleware

import (
	"github.com/allentom/haruka"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

type JWTMiddleware struct {
	Logger *logrus.Logger
	Option *NewJWTMiddlewareOption
}

func (m *JWTMiddleware) OnRequest(ctx *haruka.Context) {
	if m.Option.JWTKey == nil {
		return
	}
	tokenString := ""
	if m.Option.ReadTokenString != nil {
		tokenString = m.Option.ReadTokenString(ctx)
	} else if len(m.Option.HeaderLookUp) != 0 {
		tokenString = ctx.Request.Header.Get(m.Option.HeaderLookUp)
	} else if len(m.Option.ParamLookUp) != 0 {
		tokenString = ctx.GetQueryString(m.Option.ParamLookUp)
	}
	if len(tokenString) == 0 {
		return
	}
	var claims jwt.Claims
	if m.Option.GetClaims != nil {
		claims = m.Option.GetClaims()
	}
	claims = &jwt.StandardClaims{}
	parseToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return m.Option.JWTKey, nil
	})
	if err != nil {
		return
	}
	ctx.Param["token"] = parseToken
	ctx.Param["claims"] = claims
}

type NewJWTMiddlewareOption struct {
	ReadTokenString func(ctx *haruka.Context) string
	HeaderLookUp    string
	ParamLookUp     string
	GetClaims       func() jwt.Claims
	JWTKey          []byte
}

func NewJWTMiddleware(option *NewJWTMiddlewareOption) *JWTMiddleware {
	return &JWTMiddleware{
		Logger: logrus.New(),
		Option: option,
	}
}
