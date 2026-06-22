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

func (c *Client) post(path string, payload map[string]string) (int, []byte, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.publishableKey)

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
