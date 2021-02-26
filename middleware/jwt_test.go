package middleware

import (
	"github.com/allentom/haruka"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

var superkey = "superkey"

func createTestToken() (string, error) {
	claims := &jwt.StandardClaims{
		Id: "testuser1",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(superkey))
	if err != nil {
		return "", err
	}
	return ss, nil
}
func TestJWTMiddleware_OnRequest(t *testing.T) {
	options := &NewJWTMiddlewareOption{
		HeaderLookUp: "CustomAuthHeader",
		JWTKey:       []byte(superkey),
	}
	mid := NewJWTMiddleware(options)
	request, err := http.NewRequest("GET", "/test", strings.NewReader(""))
	if err != nil {
		t.Error(err)
		return
	}
	tokenStr, err := createTestToken()
	if err != nil {
		t.Error(err)
		return
	}
	request.Header.Set("CustomAuthHeader", tokenStr)
	context := &haruka.Context{
		Request: request,
		Param:   map[string]interface{}{},
	}
	mid.OnRequest(context)

	parseClaims := context.Param["claims"].(*jwt.StandardClaims)
	assert.Equal(t, "testuser1", parseClaims.Id)
}
