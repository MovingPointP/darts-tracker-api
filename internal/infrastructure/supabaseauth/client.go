// Package supabaseauth はSupabase AuthのREST APIへのプロキシクライアント。
// 本サーバーはサインアップ・ログイン・トークン更新の実処理をSupabase Authに委譲し、
// レスポンス(access_token/refresh_token等)をそのままクライアントへ中継する。
package supabaseauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	baseURL        string
	publishableKey string
	httpClient     *http.Client
}

// コンストラクタ
func NewClient() (*Client, error) {
	baseURL := os.Getenv("SUPABASE_URL")
	publishableKey := os.Getenv("SUPABASE_PUBLISHABLE_KEY")
	if baseURL == "" || publishableKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_PUBLISHABLE_KEY must be set")
	}
	return &Client{
		baseURL:        baseURL,
		publishableKey: publishableKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// SignUp はSupabase Authへユーザー登録をプロキシする。
func (c *Client) SignUp(email, password string) (statusCode int, body []byte, err error) {
	return c.post("/auth/v1/signup", map[string]string{"email": email, "password": password})
}

// SignInWithPassword はSupabase Authへpassword grantでのログインをプロキシする。
func (c *Client) SignInWithPassword(email, password string) (statusCode int, body []byte, err error) {
	return c.post("/auth/v1/token?grant_type=password", map[string]string{"email": email, "password": password})
}

// RefreshToken はSupabase Authへrefresh_token grantでのトークン更新をプロキシする。
func (c *Client) RefreshToken(refreshToken string) (statusCode int, body []byte, err error) {
	return c.post("/auth/v1/token?grant_type=refresh_token", map[string]string{"refresh_token": refreshToken})
}

// RequestPasswordRecovery はSupabase Authへパスワードリセットメールの送信をプロキシする。
// redirectTo にはメール内リンクの遷移先(フロントのreset-passwordページ)を指定する。
// Supabaseのリダイレクト許可リスト(Redirect URLs)に登録されている必要がある。
func (c *Client) RequestPasswordRecovery(email, redirectTo string) (statusCode int, body []byte, err error) {
	path := "/auth/v1/recover"
	if redirectTo != "" {
		path += "?redirect_to=" + url.QueryEscape(redirectTo)
	}
	return c.post(path, map[string]string{"email": email})
}

// UpdatePassword はリカバリーセッションのアクセストークンを使い、ユーザーのパスワードを更新する。
// アクセストークンはメールのリンク経由で発行されたリカバリーセッションのものを渡す。
func (c *Client) UpdatePassword(accessToken, password string) (statusCode int, body []byte, err error) {
	return c.doRequest(http.MethodPut, "/auth/v1/user", accessToken, map[string]string{"password": password})
}

// GetUser はアクセストークンの持ち主のユーザー情報(メールアドレス等)を取得する。
func (c *Client) GetUser(accessToken string) (statusCode int, body []byte, err error) {
	return c.doRequest(http.MethodGet, "/auth/v1/user", accessToken, nil)
}

func (c *Client) post(path string, payload map[string]string) (int, []byte, error) {
	return c.doRequest(http.MethodPost, path, "", payload)
}

// doRequest はSupabase Auth REST APIへJSONリクエストを送り、ステータスとレスポンスボディを返す。
// bearerToken が空でなければ Authorization ヘッダーを付与する(ユーザー操作系エンドポイント用)。
// payload が nil の場合はボディ無しで送る(GET等)。
func (c *Client) doRequest(method, path, bearerToken string, payload map[string]string) (int, []byte, error) {
	var reqBody io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.publishableKey)
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to call supabase auth: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read supabase auth response: %w", err)
	}
	return resp.StatusCode, respBody, nil
}
