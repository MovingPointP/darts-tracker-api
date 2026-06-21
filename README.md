# darts-tracker-api

Go + Gin + Supabase で構築するダーツ得点記録アプリのバックエンド REST API。

01Game(1ラウンド平均点)・クリケット(1ラウンド平均マーク数)・COUNTUP(合計得点)の3種目を、ゲーム単位の集計値のみで記録する。01Gameとクリケットについては、DARTSLIVE風の換算表をもとにレーティングを自動算出する。

設計の経緯・詳細は [docs/design.md](docs/design.md) を参照。

## 技術スタック

| カテゴリ | 技術 |
|----------|------|
| 言語 | [Go](https://go.dev/) |
| Webフレームワーク | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io/) |
| データベース | [Supabase](https://supabase.com/) (PostgreSQL) |
| 認証 | Supabase Auth(本APIがプロキシし、JWTはJWKS/共有シークレットで検証のみ行う) |
| APIドキュメント | [swaggo/swag](https://github.com/swaggo/swag) (OpenAPI) |
| 環境変数 | [godotenv](https://github.com/joho/godotenv) |
| コンテナ | Docker |
| デプロイ | [Render](https://render.com/) |

## アーキテクチャ

[go-task-api](https://github.com/MovingPointP/go-task-api) と同様のクリーンアーキテクチャを採用。

```
darts-tracker-api/
├── cmd/server/          # エントリーポイント
├── docs/                # design.md(設計ドキュメント)、swag生成のOpenAPIドキュメント
├── internal/
│   ├── domain/
│   │   ├── entity/      # エンティティ(ビジネスオブジェクト)
│   │   └── repository/  # リポジトリインターフェース
│   ├── usecase/
│   │   ├── game_record_usecase.go
│   │   └── rating/      # レーティング換算ロジック
│   ├── infrastructure/  # DB・Supabase Auth連携の実装
│   └── handler/         # HTTPハンドラ、JWT検証ミドルウェア
```

## セットアップ

```bash
cp .env.example .env
# .env に DATABASE_URL / SUPABASE_URL / SUPABASE_ANON_KEY 等を設定
go run cmd/server/main.go
```

または Docker:

```bash
docker compose up
```

## API概要

| メソッド | パス | 認証 | 内容 |
|---|---|---|---|
| POST | /api/v1/auth/signup | 不要 | Supabase Authへプロキシ |
| POST | /api/v1/auth/login | 不要 | Supabase Authへプロキシ |
| POST | /api/v1/auth/refresh | 不要 | Supabase Authへプロキシ |
| GET | /api/v1/records | 必要 | 記録一覧(種目フィルタ可) |
| POST | /api/v1/records | 必要 | 記録作成(レーティング自動計算) |
| PUT | /api/v1/records/:id | 必要 | 記録更新 |
| DELETE | /api/v1/records/:id | 必要 | 記録削除 |

詳細は起動後 `/swagger/index.html` を参照。
