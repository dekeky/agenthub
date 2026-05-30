package hub

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"github.com/dekeky/rssmanager/pkg/ginx"
	"github.com/gin-gonic/gin"
)

func (hr *Router) uploadAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if hr.uploadToken == "" {
			ginx.NewRender(c, http.StatusUnauthorized).Err(fmt.Errorf("upload is disabled: AGENTHUB_UPLOAD_TOKEN is not configured"))
			c.Abort()
			return
		}

		got := bearerToken(c.GetHeader("Authorization"))
		if got == "" {
			got = strings.TrimSpace(c.GetHeader("X-Upload-Token"))
		}
		if !tokenEqual(got, hr.uploadToken) {
			ginx.NewRender(c, http.StatusUnauthorized).Err(fmt.Errorf("invalid upload token"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

func tokenEqual(got, want string) bool {
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
