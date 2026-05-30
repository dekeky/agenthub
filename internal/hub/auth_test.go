package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		token      string
		authHeader string
		uploadHdr  string
		wantStatus int
	}{
		{
			name:       "missing configured token",
			token:      "",
			authHeader: "Bearer secret",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing client token",
			token:      "secret",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			token:      "secret",
			authHeader: "Bearer wrong",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid bearer token",
			token:      "secret",
			authHeader: "Bearer secret",
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid x-upload-token",
			token:      "secret",
			uploadHdr:  "secret",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			hr := &Router{uploadToken: tt.token}
			r.POST("/upload", hr.uploadAuth(), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/upload", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.uploadHdr != "" {
				req.Header.Set("X-Upload-Token", tt.uploadHdr)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
