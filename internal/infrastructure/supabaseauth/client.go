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
	// secretKey はSupabaseのservice_role(sb_secret_...)。管理者API(ユーザー削除等)にのみ使う。
	// 全権限を持つため、外部へは絶対に露出させない。退会機能を使わない場合は未設定でも起動できる。
	secretKey  string
	httpClient *http.Client
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
		secretKey:      os.Getenv("SUPABASE_SECRET_KEY"),
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

// UpdateEmail はメールアドレスの変更を申請する。
// 実際に変更されるのは確認メールのリンクが踏まれた後で、この時点では反映されない
// (Supabaseの Secure email change が有効な場合、新旧両方のアドレスでの確認が必要)。
// redirectTo には確認リンクの遷移先を指定し、Supabaseのリダイレクト許可リストに登録しておく。
func (c *Client) UpdateEmail(accessToken, email, redirectTo string) (statusCode int, body []byte, err error) {
	path := "/auth/v1/user"
	if redirectTo != "" {
		path += "?redirect_to=" + url.QueryEscape(redirectTo)
	}
	return c.doRequest(http.MethodPut, path, accessToken, map[string]string{"email": email})
}

// DeleteUser はSupabaseの管理者APIでユーザーを完全に削除する(退会)。
// service_role(secretKey)が必要。未設定の場合はエラーを返す。
func (c *Client) DeleteUser(userID string) (statusCode int, body []byte, err error) {
	if c.secretKey == "" {
		return 0, nil, fmt.Errorf("SUPABASE_SECRET_KEY is not configured")
	}

	req, err := http.NewRequest(
		http.MethodDelete,
		c.baseURL+"/auth/v1/admin/users/"+url.PathEscape(userID),
		nil,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to build request: %w", err)
	}
	// 管理者APIではapikey/Authorizationの両方にsecret key(service_role相当)を用いる。
	req.Header.Set("apikey", c.secretKey)
	req.Header.Set("Authorization", "Bearer "+c.secretKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to call supabase admin api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read supabase admin api response: %w", err)
	}
	return resp.StatusCode, respBody, nil
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
