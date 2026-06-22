package handler

import (
	"net/http"

	"github.com/MovingPointP/darts-tracker-api/internal/infrastructure/supabaseauth"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	supabaseClient *supabaseauth.Client
}

func NewAuthHandler(supabaseClient *supabaseauth.Client) *AuthHandler {
	return &AuthHandler{supabaseClient: supabaseClient}
}

// サインアップのリクエストボディ
type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// ログインのリクエストボディ
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// トークン更新のリクエストボディ
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary     サインアップ
// @Description Supabase Authへプロキシしてユーザー登録する。レスポンスはSupabaseのものをそのまま返す
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body SignUpRequest true "登録情報"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /auth/signup [post]
func (h *AuthHandler) SignUp(ctx *gin.Context) {
	var req SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, body, err := h.supabaseClient.SignUp(req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign up"})
		return
	}
	ctx.Data(status, "application/json", body)
}

// @Summary     ログイン
// @Description Supabase Authへプロキシしてログインする。レスポンスはSupabaseのものをそのまま返す
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body LoginRequest true "ログイン情報"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /auth/login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, body, err := h.supabaseClient.SignInWithPassword(req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}
	ctx.Data(status, "application/json", body)
}

// @Summary     トークン更新
// @Description Supabase Authへプロキシしてアクセストークンを更新する。レスポンスはSupabaseのものをそのまま返す
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RefreshRequest true "リフレッシュトークン"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /auth/refresh [post]
func (h *AuthHandler) Refresh(ctx *gin.Context) {
	var req RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, body, err := h.supabaseClient.RefreshToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh token"})
		return
	}
	ctx.Data(status, "application/json", body)
}
