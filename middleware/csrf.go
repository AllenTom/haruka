package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/allentom/haruka"
)

// CSRFMiddleware implements the double-submit cookie pattern.
//
// On safe methods (GET/HEAD/OPTIONS) it ensures a token cookie exists
// and exposes the token via ctx.Param["csrfToken"] so templates can
// embed it. Unsafe methods must present the same token through the
// configured header or form field; mismatches abort with 403.
type CSRFMiddleware struct {
	// CookieName defaults to "csrf_token".
	CookieName string
	// HeaderName defaults to "X-CSRF-Token".
	HeaderName string
	// FieldName checks multipart/form bodies when set (default empty).
	FieldName string
	// Secure marks the cookie HTTPS-only (default true).
	Secure *bool
	// Skip allows bypassing verification for selected requests
	// (e.g. token-authenticated API calls).
	Skip func(ctx *haruka.Context) bool

	initOnce sync.Once
}

func (m *CSRFMiddleware) cookieName() string {
	if m.CookieName != "" {
		return m.CookieName
	}
	return "csrf_token"
}

func (m *CSRFMiddleware) headerName() string {
	if m.HeaderName != "" {
		return m.HeaderName
	}
	return "X-CSRF-Token"
}

func (m *CSRFMiddleware) secure() bool {
	m.initOnce.Do(func() {
		if m.Secure == nil {
			v := true
			m.Secure = &v
		}
	})
	return *m.Secure
}

func generateCSRFToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

func (m *CSRFMiddleware) OnRequest(ctx *haruka.Context) {
	if m.Skip != nil && m.Skip(ctx) {
		return
	}
	cookie, err := ctx.Request.Cookie(m.cookieName())
	token := ""
	if err == nil {
		token = cookie.Value
	}
	if token == "" {
		token = generateCSRFToken()
		if token == "" {
			return
		}
		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:     m.cookieName(),
			Value:    token,
			Path:     "/",
			Secure:   m.secure(),
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		})
	}
	ctx.Param["csrfToken"] = token

	if isSafeMethod(ctx.Request.Method) {
		return
	}

	presented := ctx.Request.Header.Get(m.headerName())
	if presented == "" && m.FieldName != "" && ctx.Request.Form != nil {
		if values := ctx.Request.Form[m.FieldName]; len(values) > 0 {
			presented = values[0]
		}
	}
	if presented == "" ||
		subtle.ConstantTimeCompare([]byte(presented), []byte(token)) != 1 {
		ctx.AbortWithError(haruka.ErrForbidden("CSRF token mismatch"))
	}
}
