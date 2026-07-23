package handler

import (
	"net/http"
	"strings"

	"github.com/MovingPointP/darts-tracker-api/internal/infrastructure/supabaseauth"
	"github.com/MovingPointP/darts-tracker-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	supabaseClient    *supabaseauth.Client
	gameRecordUsecase usecase.GameRecordUsecase
}

func NewAuthHandler(supabaseClient *supabaseauth.Client, gameRecordUsecase usecase.GameRecordUsecase) *AuthHandler {
	return &AuthHandler{supabaseClient: supabaseClient, gameRecordUsecase: gameRecordUsecase}
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

// パスワードリセットメール送信のリクエストボディ
type RecoverRequest struct {
	Email      string `json:"email" binding:"required,email"`
	RedirectTo string `json:"redirect_to"`
}

// パスワード再設定のリクエストボディ
type ResetPasswordRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
}

// メールアドレス変更のリクエストボディ
type ChangeEmailRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	RedirectTo  string `json:"redirect_to"`
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

// @Summary     パスワードリセットメール送信
// @Description Supabase Authへプロキシしてパスワードリセット用のメールを送信する。メールアドレスの存在有無を漏らさないよう、Supabaseは登録有無に関わらず成功を返す
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RecoverRequest true "リセット対象のメールアドレスと遷移先URL"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /auth/recover [post]
func (h *AuthHandler) RequestPasswordReset(ctx *gin.Context) {
	var req RecoverRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, body, err := h.supabaseClient.RequestPasswordRecovery(req.Email, req.RedirectTo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request password reset"})
		return
	}
	ctx.Data(status, "application/json", body)
}

// @Summary     パスワード再設定
// @Description リカバリーセッションのアクセストークンを使い、Supabase Authへプロキシして新しいパスワードを設定する
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body ResetPasswordRequest true "リカバリーのアクセストークンと新パスワード"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, body, err := h.supabaseClient.UpdatePassword(req.AccessToken, req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}
	ctx.Data(status, "application/json", body)
}

// @Summary     メールアドレス変更の申請
// @Description Supabase Authへプロキシしてメールアドレス変更の確認メールを送信する。確認リンクが踏まれるまで実際のアドレスは変更されない
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body ChangeEmailRequest true "アクセストークンと新しいメールアドレス"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /auth/change-email [post]
func (h *AuthHandler) ChangeEmail(ctx *gin.Context) {
	var req ChangeEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, body, err := h.supabaseClient.UpdateEmail(req.AccessToken, req.Email, req.RedirectTo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change email"})
		return
	}
	ctx.Data(status, "application/json", body)
}

// @Summary     ログイン中ユーザーの情報取得
// @Description アクセストークンの持ち主のユーザー情報(メールアドレス・登録日等)をSupabase Authから取得する
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} map[string]string
// @Router      /auth/me [get]
func (h *AuthHandler) Me(ctx *gin.Context) {
	// 認証ミドルウェアで検証済みのアクセストークンをそのままSupabaseへ中継する。
	accessToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if accessToken == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
		return
	}

	status, body, err := h.supabaseClient.GetUser(accessToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	ctx.Data(status, "application/json", body)
}

// @Summary     アカウント削除(退会)
// @Description ログイン中ユーザーの全記録を削除し、Supabaseの認証ユーザーも管理者APIで削除する。不可逆
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     204
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/account [delete]
func (h *AuthHandler) DeleteAccount(ctx *gin.Context) {
	userID := ctx.MustGet("UserID").(string)

	// 先にSupabaseの認証ユーザーを削除する(service_roleが必要)。
	// こちらの方が失敗しやすい(キー設定ミス等)ため先に実行し、失敗時は記録を消さずに
	// 中断して安全に再試行できるようにする(記録先行で消すと「記録だけ消えてアカウントが残る」半壊になる)。
	status, body, err := h.supabaseClient.DeleteUser(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete auth user"})
		return
	}
	if status < 200 || status >= 300 {
		ctx.Data(status, "application/json", body)
		return
	}

	// 認証ユーザー削除に成功したら、アプリのデータ(記録)を削除する。
	// auth.usersへの外部キーが無くカスケードされないため明示的に削除する。
	if err := h.gameRecordUsecase.DeleteAllByUser(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user records"})
		return
	}

	ctx.Status(http.StatusNoContent)
}
