# デプロイ手順

## 環境別起動方法

### ローカル開発

```bash
cp .env.example .env
# .env を編集して接続情報を設定

go run cmd/api/main.go
# → http://localhost:8081
```

### Docker Compose（ローカル統合テスト）

```bash
cp .env.example .env
# MONGO_PASSWORD を設定

docker compose up -d
# → http://localhost:8081

docker compose logs -f api  # ログ確認
docker compose down         # 停止
```

### 本番（手動デプロイ）

```bash
# 1. イメージビルド
docker build -t idol-api:$(git rev-parse --short HEAD) .

# 2. 環境変数を設定（Secrets 管理推奨）
export MONGODB_URI="mongodb+srv://..."
export MONGODB_DATABASE="idol_database"
export SERVER_PORT="8080"
export GIN_MODE="release"
export CORS_ALLOWED_ORIGINS="https://your-frontend.example.com"
export ADMIN_API_KEY="$(openssl rand -hex 32)"
export WRITE_API_KEY="$(openssl rand -hex 32)"

# 3. コンテナ起動
docker run -d \
  --name idol-api \
  -p 8080:8080 \
  -e MONGODB_URI \
  -e MONGODB_DATABASE \
  -e SERVER_PORT \
  -e GIN_MODE \
  -e CORS_ALLOWED_ORIGINS \
  -e ADMIN_API_KEY \
  -e WRITE_API_KEY \
  idol-api:$(git rev-parse --short HEAD)
```

## CI/CD パイプライン

`.github/workflows/ci.yml` が以下を自動実行します。

| ジョブ | トリガー | 内容 |
|---|---|---|
| build-and-test | push/PR | go vet / go build / go test |
| boundary-check | push/PR | レイヤ依存方向違反チェック |
| openapi-contract | push/PR | Swagger spec の同期チェック・バリデーション |
| docker-build | push/PR | Docker イメージビルド確認 |

### GHCR への自動 publish

`.github/workflows/docker-publish.yml` により、`main` ブランチへのマージで GHCR にイメージが自動 publish されます。

- イメージ名: `ghcr.io/kuro48/idol-database:main`
- タグ付きリリース（`v*.*.*`）では `ghcr.io/kuro48/idol-database:<version>` も同時に publish されます。

## グレースフルシャットダウン

SIGTERM を受信すると、以下の順でシャットダウンします。

1. インフライト Webhook 配信が完了するまで待機
2. インフライトジョブ（非同期ジョブ）が完了するまで待機
3. HTTP サーバーを最大 30 秒で停止

コンテナオーケストレーター（Kubernetes など）や `docker stop` はデフォルトで SIGTERM を送信するため、追加設定なしでグレースフルシャットダウンが機能します。

## ヘルスチェック

```bash
curl http://localhost:8081/health
# → {"status":"ok"}
```

## ロールバック手順

```bash
# 前バージョンのイメージタグを確認
docker images idol-api

# 旧バージョンを起動
docker stop idol-api
docker run -d --name idol-api <旧タグのイメージ> ...
```

## リリースチェックリスト

- [ ] `go test ./...` がパス
- [ ] `go vet ./...` がパス
- [ ] CI の全ジョブがグリーン
- [ ] `.env` の `GIN_MODE=release` を確認
- [ ] `GIN_MODE=release` で `/swagger/index.html` が 404 になること
- [ ] `ADMIN_API_KEY` が本番用の値か確認
- [ ] `WRITE_API_KEY` が本番用の値か確認
- [ ] `ADMIN_API_KEY` / `WRITE_API_KEY` ともに `openssl rand -hex 32` で生成したキーを使用していること
- [ ] `CORS_ALLOWED_ORIGINS` が本番フロントエンドURLか確認
- [ ] MongoDB接続先が本番DBか確認
- [ ] ヘルスチェック `/health` が正常応答

## 障害時対応

1. ログ確認: `docker logs idol-api --tail 100`
2. ヘルスチェック: `curl /health`
3. 問題が解決しない場合: ロールバック実施
4. MongoDB 接続エラーの場合: `MONGODB_URI` を確認
