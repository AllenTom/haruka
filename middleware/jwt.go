package middleware

import (
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/allentom/haruka"
	"github.com/sirupsen/logrus"
)

type JWTMiddleware struct {
	Logger *logrus.Logger
	Option *NewJWTMiddlewareOption
}

// NewJWTMiddlewareOption configures token extraction.
//
// Token lookup order:
//  1. ReadTokenString callback
//  2. HeaderLookUp header name
//  3. ParamLookUp query string key
type NewJWTMiddlewareOption struct {
	ReadTokenString func(ctx *haruka.Context) string
	HeaderLookUp    string
	ParamLookUp     string
	// GetClaims returns the claims struct to unmarshal into.
	// Defaults to *jwt.RegisteredClaims.
	GetClaims func() jwt.Claims
	JWTKey    []byte
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
	} else {
		claims = &jwt.RegisteredClaims{}
	}
	parseToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return m.Option.JWTKey, nil
	})
	if err != nil {
		if m.Logger != nil {
			m.Logger.WithError(err).Debug("jwt parse failed")
		}
		return
	}
	ctx.Param["token"] = parseToken
	ctx.Param["claims"] = claims
}

func NewJWTMiddleware(option *NewJWTMiddlewareOption) *JWTMiddleware {
	return &JWTMiddleware{
		Logger: logrus.New(),
		Option: option,
	}
}
