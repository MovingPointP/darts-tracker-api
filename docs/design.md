# ダーツ得点記録アプリ 実装計画

## Context

ダーツの自己練習記録を残すための個人用アプリを新規構築する。01Game(平均点)・クリケット(平均マーク数)・COUNTUP(合計得点)の3種目を、投/ラウンド単位ではなく1ゲームの集計値のみで記録し、01Gameとクリケットについては記録からDARTSLIVE風のレーティングを算出して保存する。

既存の `/home/pointp/Develop/go-gin-tutorial/go-task-api` (Go+Gin+GORM+Supabase+Renderのクリーンアーキテクチャ構成)を踏襲しつつ、認証はSupabase Authに委譲し、フロントはNext.js+TypeScript+Mantineで構築する。フロント・バックは別リポジトリ・別ホスティング(Vercel/Render)。

過去の対話で以下を確定済み:
- 個人アカウント制、Supabase + Supabase Auth、認証はGoバックエンドがプロキシ(Supabaseのトークンをそのまま中継し、GoはJWKS検証のみ行う)
- レーティングは01Game(PPR)とクリケット(MPR)それぞれ独立に算出、統合しない。非公式換算表(dartsmap.com)のバンドを参考に、バンド内線形補間で小数2桁まで算出
- 記録項目は最小構成: 日付・種目・値(PPR/MPR/合計点)・レーティング(該当種目のみ、nullable)
- 開発用/本番用でSupabaseプロジェクトを分離
- フロントMVPは4画面: ログイン/サインアップ、記録入力、記録一覧(編集・削除可)、レーティング推移グラフ
- フロントの技術選定: Mantine(UI) + @mantine/form + zod + @mantine/charts(recharts) + SWR

## レーティング換算テーブル(両者共通ロジック)

`dartsmap.com`のDARTSLIVE 80%スタッツ表をもとに、以下のバンド配列をPPR用・MPR用それぞれ用意する。

```go
type ratingBand struct {
    Min, Max float64 // Maxは次バンドの下限と一致する半開区間 [Min, Max)
    Rating   float64 // このバンドの下限での基準レーティング
}
```

例(01Game/PPR、抜粋): `{0,40,1}, {40,45,2}, {45,50,3}, {50,55,4}, ..., {95,102,13}, ..., {130, +Inf, 18}`
例(Cricket/MPR、抜粋): `{0,1.3,1}, {1.3,1.5,2}, ..., {3.5,3.75,13}, ..., {4.75, +Inf, 18}`

補間式: `rating = band.Rating + (value - band.Min) / (band.Max - band.Min)`、小数2桁に丸める。最上位バンド(上限なし)はそのバンドの`Rating`値(18.00)で固定(青天井で伸ばさない)。最下位バンドの下限(0)は値0のとき1.00になる。

実装は `internal/usecase/rating` パッケージ(または `pkg/rating`)に切り出し、`CalculatePPRRating(ppr float64) float64` と `CalculateMPRRating(mpr float64) float64` の2関数 + 表データを置く。COUNTUPはレーティング計算対象外。

## バックエンド: darts-tracker-api (新規リポジトリ)

go-task-apiの規約をそのまま踏襲する箇所:
- entity層: `uint` autoincrement ID、GORMタグ直書き、ドメインエラーは `var ErrXxx = errors.New(...)` をentity内に定義
- repository層: `context.Context`を使わないシンプルなインターフェース、Not Foundは`(nil, nil)`
- usecase層: `interface` + 非公開struct + `New*`コンストラクタDI、エラーは`fmt.Errorf("...: %w", err)`でラップ、更新/削除前に所有権チェック
- infrastructure層: GORM + `db.AutoMigrate(...)`、`gorm.ErrRecordNotFound`を`(nil, nil)`に変換
- handler層: `ShouldBindJSON`+`binding`タグ、エラーは`gin.H{"error": "..."}`、ステータスコード使い分け
- マルチステージDockerfile、`swag init`によるswagger生成、`godotenv`での`.env`読込

変更が必要な箇所:
- `pkg/jwt`の`GenerateToken`(自前JWT発行)は廃止。代わりに**JWT検証のみ**を行うミドルウェアに置き換え、SupabaseのJWKS (or 共有シークレット、プロジェクト作成後に方式確定)で署名検証し、`sub`クレーム(UUID文字列)を`UserID`としてコンテキストにセットする
- `auth_handler.go`はSupabase Auth APIへのプロキシに役割変更: `/auth/signup` `/auth/login` `/auth/refresh` がSupabaseの`/auth/v1/...`エンドポイントを呼び出し、レスポンス(access_token/refresh_token)をそのまま返す
- `UserID`の型は`uint`ではなく`string`(UUID)。Supabaseの`auth.users.id`がUUIDのため、`GameRecord.UserID`もUUID文字列で保持する

### ディレクトリ構成

```
darts-tracker-api/
├── cmd/server/main.go
├── internal/
│   ├── domain/
│   │   ├── entity/game_record.go
│   │   └── repository/game_record_repository.go
│   ├── usecase/
│   │   ├── game_record_usecase.go
│   │   └── rating/rating.go            # 換算テーブル+補間ロジック
│   ├── infrastructure/
│   │   ├── db/db.go                    # GORM接続+AutoMigrate
│   │   ├── persistence/game_record_repository_impl.go
│   │   └── supabaseauth/client.go      # Supabase Auth APIプロキシ用クライアント + JWKS検証
│   └── handler/
│       ├── game_record_handler.go
│       ├── auth_handler.go
│       └── middleware/auth.go          # JWT検証ミドルウェア(発行ではなく検証)
├── docs/                                # swag生成
├── Dockerfile
├── docker-compose.yml                   # ローカルでもリモートSupabaseのDATABASE_URLに接続(go-task-api踏襲)
├── .env.example                         # PORT, DATABASE_URL, SUPABASE_URL, SUPABASE_ANON_KEY, SUPABASE_JWT_SECRET(or JWKS URL)
└── go.mod
```

### entity (`game_record.go`)

```go
type GameType string
const (
    GameType01Game   GameType = "01game"
    GameTypeCricket  GameType = "cricket"
    GameTypeCountUp  GameType = "countup"
)

type GameRecord struct {
    ID       uint      `json:"id" gorm:"primaryKey"`
    UserID   string    `json:"user_id" gorm:"not null;index;type:uuid"`
    GameType GameType  `json:"game_type" gorm:"not null"`
    Value    float64   `json:"value" gorm:"not null"`     // PPR / MPR / 合計点
    Rating   *float64  `json:"rating"`                    // 01game/cricketのみ、countupはnull
    PlayedAt time.Time `json:"played_at" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
var ErrGameRecordNotFound = errors.New("game record not found")
```

### APIエンドポイント

| メソッド | パス | 認証 | 内容 |
|---|---|---|---|
| POST | /api/v1/auth/signup | 不要 | Supabase Authへプロキシ |
| POST | /api/v1/auth/login | 不要 | Supabase Authへプロキシ(password grant) |
| POST | /api/v1/auth/refresh | 不要 | Supabase Authへプロキシ(refresh grant) |
| GET | /api/v1/records | 必要 | 一覧(`?game_type=` / `?from=YYYY-MM-DD` / `?to=YYYY-MM-DD` / `?limit=` / `?offset=`)。レスポンスは`{records, total, limit, offset}` |
| POST | /api/v1/records | 必要 | 作成(01game/cricketはサーバー側でレーティング自動計算) |
| PUT | /api/v1/records/:id | 必要 | 更新(値変更時はレーティング再計算) |
| DELETE | /api/v1/records/:id | 必要 | 削除 |
| GET | /api/v1/stats/ratings | 必要 | 日別平均レーティング(`?game_type=01game\|cricket`必須)。DBで`GROUP BY date / AVG(rating)`して返す。グラフ表示用 |

CORSはVercelの本番/プレビュードメインを許可リストに追加(`gin-contrib/cors`を新規導入)。

## フロントエンド: darts-tracker-web (新規リポジトリ)

```
darts-tracker-web/
├── app/
│   ├── login/page.tsx
│   ├── signup/page.tsx
│   ├── records/page.tsx        # 一覧+編集+削除、種目フィルタ
│   ├── records/new/page.tsx    # 入力フォーム(or Mantine Modal化)
│   ├── stats/page.tsx          # レーティング推移グラフ(01Game/Cricket別)
│   └── layout.tsx              # MantineProvider, AppShellナビ
├── lib/
│   ├── api-client.ts           # fetchラッパー、Authorizationヘッダー付与、401時にrefresh
│   ├── auth-context.tsx        # access/refresh tokenの保持(localStorage)、useAuthフック
│   └── fetcher.ts              # SWR用fetcher
├── components/
│   ├── RecordForm.tsx          # @mantine/form + zod
│   ├── RecordTable.tsx
│   └── RatingChart.tsx         # @mantine/charts
├── types/record.ts             # バックエンドのJSONと対応する型
└── .env.local.example          # NEXT_PUBLIC_API_BASE_URL
```

トークン保存はlocalStorageを使う(個人利用前提のため、httpOnly Cookie+クロスオリジン設定の複雑さを避ける実用的な判断。将来複数ユーザー化する際はCookie方式への切替を検討)。

## 実装順序

1. **Supabase準備(ユーザー作業)**: 開発用・本番用のSupabaseプロジェクトを2つ作成し、`DATABASE_URL` / `SUPABASE_URL` / `SUPABASE_ANON_KEY` / JWT検証用シークレットまたはJWKS URLを取得してもらう(アカウント操作はこちらでは実行できないため)
2. **バックエンド scaffold**: go.mod, Dockerfile, docker-compose.yml, .env.example, ディレクトリ作成
3. **GameRecord CRUD**: entity → repository → usecase → infrastructure(AutoMigrate) → handler、swaggerアノテーション付与
4. **レーティング計算**: `rating`パッケージ実装+ユニットテスト(go testing、境界値・補間を検証)
5. **認証**: Supabase Authプロキシハンドラ(signup/login/refresh) + JWT検証ミドルウェア、`records`ルートに適用
6. **バックエンド動作確認**: ローカルでswagger UIまたはcurlでsignup→login→records CRUDを一通り確認
7. **フロントエンド scaffold**: `create-next-app`(TypeScript) + Mantine導入(`@mantine/core` `@mantine/form` `@mantine/charts` `@mantine/hooks`) + SWR
8. **認証画面**: signup/loginフォーム → Goバックエンド経由でトークン取得・保存
9. **記録機能**: 入力フォーム、一覧+編集+削除、SWRでの取得・再検証
10. **グラフ画面**: `@mantine/charts`でレーティング推移を表示
11. **デプロイ設定**: Render(バックエンド、環境変数設定)、Vercel(フロントエンド、`NEXT_PUBLIC_API_BASE_URL`設定)、CORS許可ドメイン設定
12. **E2E確認**: 本番相当環境でsignup→記録追加(3種目)→レーティング表示→編集・削除の一連の流れを確認

## 検証方法

- バックエンド: `go test ./...` でレーティング計算のユニットテスト(既知のPPR/MPR値に対する期待レーティング値を複数ケース用意)、swagger UIまたはcurlでAPIの手動疎通確認
- フロントエンド: `npm run build`で型エラーなしを確認、ローカルで実際にサインアップ〜記録登録〜グラフ表示までブラウザ操作で確認
- 認証フロー: Goバックエンドのsignup/loginが実際にSupabaseプロジェクトにユーザーを作成し、返ってきたトークンで`/api/v1/records`にアクセスできることを確認

## 解決済み: JWT署名方式

新規作成したSupabaseプロジェクトはJWT Signing Keys方式(ES256非対称鍵)がデフォルトで、ダッシュボードに共有シークレット(HS256)の設定項目が存在しなかった。そのため、`internal/handler/middleware/auth.go`は`SUPABASE_JWT_SECRET`を使うHS256検証ではなく、`{SUPABASE_URL}/auth/v1/.well-known/jwks.json`から公開鍵を取得して検証するJWKS方式(`github.com/MicahParks/keyfunc/v3`を使用)で実装している。`.env`に`SUPABASE_JWT_SECRET`の設定は不要。

## 解決済み: APIキーの名称

同様にAPIキー体系も刷新されており、`anon key`は**Publishable key**(`sb_publishable_...`)に名称・形式が変わっている(`service_role`は`Secret key`に対応)。Supabase Auth APIのsignup/login/refresh呼び出しは公式ドキュメントのcURL例でもPublishable keyを使う想定のため、`SUPABASE_ANON_KEY`という環境変数名は`SUPABASE_PUBLISHABLE_KEY`に変更した。Secret keyは管理者操作用のため本アプリでは使用しない。
