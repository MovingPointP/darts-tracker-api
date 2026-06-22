package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// supabaseClaims はSupabase Authが発行するアクセストークンのクレーム。
// "sub"クレームがSupabaseのauth.users.id(UUID文字列)に対応する。
type supabaseClaims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

// NewAuthMiddleware はSupabase AuthのJWKSエンドポイントから公開鍵を取得し、
// それをもとに発行されたJWT(ES256/非対称鍵のJWT Signing Keys方式)を検証する
// ミドルウェアを構築する(発行は行わない)。
// keyfuncが鍵をバックグラウンドで自動更新するため、ミドルウェアはプロセス起動時に
// 一度だけ構築し使い回す。
func NewAuthMiddleware(ctx context.Context) (gin.HandlerFunc, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is not set")
	}
	jwksURL := supabaseURL + "/auth/v1/.well-known/jwks.json"

	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch supabase JWKS: %w", err)
	}

	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Authorization header is required"},
			)
			return
		}

		// "Bearer <token>"の形式かチェック
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Authorization header format must be Bearer {token}"},
			)
			return
		}

		claims := &supabaseClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, k.Keyfunc)
		if err != nil || !token.Valid || claims.Sub == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid or expired token"},
			)
			return
		}

		ctx.Set("UserID", claims.Sub)
		ctx.Next()
	}, nil
}
