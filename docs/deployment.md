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
  idol-api:$(git rev-parse --short HEAD)
```

## CI/CD パイプライン

`.github/workflows/ci.yml` が以下を自動実行します。

| ジョブ | トリガー | 内容 |
|---|---|---|
| build-and-test | push/PR | go vet / go build / go test |
| boundary-check | push/PR | レイヤ依存方向違反チェック |
| docker-build | push/PR | Docker イメージビルド確認 |

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
- [ ] `ADMIN_API_KEY` が本番用の値か確認
- [ ] `CORS_ALLOWED_ORIGINS` が本番フロントエンドURLか確認
- [ ] MongoDB接続先が本番DBか確認
- [ ] ヘルスチェック `/health` が正常応答
- [ ] Swagger UI `/swagger/index.html` が正常表示

## 障害時対応

1. ログ確認: `docker logs idol-api --tail 100`
2. ヘルスチェック: `curl /health`
3. 問題が解決しない場合: ロールバック実施
4. MongoDB 接続エラーの場合: `MONGODB_URI` を確認
