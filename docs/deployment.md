# デプロイ手順

## 環境別起動方法

### ローカル開発

```bash
cp .env.example .env.local
# .env.local を編集して接続情報を設定

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
export SERVER_PORT="8081"
export GIN_MODE="release"
export CORS_ALLOWED_ORIGINS="https://your-frontend.example.com"
export WRITE_API_KEY="$(openssl rand -hex 32)"
export ADMIN_API_KEY="$(openssl rand -hex 32)"

# 3. コンテナ起動
docker run -d \
  --name idol-api \
  -p 8081:8081 \
  -e MONGODB_URI \
  -e MONGODB_DATABASE \
  -e SERVER_PORT \
  -e GIN_MODE \
  -e CORS_ALLOWED_ORIGINS \
  -e WRITE_API_KEY \
  -e ADMIN_API_KEY \
  idol-api:$(git rev-parse --short HEAD)
```

`GIN_MODE=release` では Swagger UI は無効です。API 疎通確認は `/health/live` と `/health/ready` を使用してください。

## CI/CD パイプライン

`.github/workflows/ci.yml` が以下を自動実行します。

| ジョブ | トリガー | 内容 |
|---|---|---|
| build-and-test | push/PR | go vet / go build / go test |
| boundary-check | push/PR | レイヤ依存方向違反チェック |
| docker-build | push/PR | Docker イメージビルド確認 |

`.github/workflows/docker-publish.yml` により、`main` への push と `v*.*.*` タグ push で GHCR にイメージを publish します。

## 現在のリリースフロー

現時点では、CI/GHCR 連携 + 手動デプロイを前提とします。

1. `main` にリリース対象をマージする
2. GitHub Actions の CI と Docker publish が成功していることを確認する
3. 正式リリース時は semver 形式のタグを付与する

```bash
git tag v0.1.0
git push origin v0.1.0
```

4. デプロイ先で対象イメージを pull する

```bash
docker pull ghcr.io/kuro48/idol-database:v0.1.0
```

5. 旧コンテナを停止し、新イメージで起動する
6. `/health/live` と `/health/ready` で疎通確認する

ブランチ push 時のイメージは検証用、semver タグ付きイメージは正式リリース用として扱います。

## ヘルスチェック

```bash
curl http://localhost:8081/health
# → {"status":"ok","message":"Idol API is running with DDD architecture"}

curl http://localhost:8081/health/live
# → {"status":"ok"}

curl http://localhost:8081/health/ready
# → {"status":"ok"}
```

## ロールバック手順

```bash
# 前バージョンのイメージタグを確認
docker images ghcr.io/kuro48/idol-database

# 旧バージョンを起動
docker stop idol-api
docker run -d --name idol-api ghcr.io/kuro48/idol-database:<旧タグ> ...
```

## リリースチェックリスト

- [ ] `go test ./...` がパス
- [ ] `go vet ./...` がパス
- [ ] CI の全ジョブがグリーン
- [ ] 本番環境変数の `GIN_MODE=release` を確認
- [ ] `WRITE_API_KEY` が本番用の値か確認
- [ ] `ADMIN_API_KEY` が本番用の値か確認
- [ ] `CORS_ALLOWED_ORIGINS` が本番フロントエンドURLか確認
- [ ] MongoDB接続先が本番DBか確認
- [ ] ヘルスチェック `/health/live` と `/health/ready` が正常応答
- [ ] `GIN_MODE=release` 時に Swagger UI が公開されていないことを確認

## 障害時対応

1. ログ確認: `docker logs idol-api --tail 100`
2. ヘルスチェック: `curl /health/live` と `curl /health/ready`
3. 問題が解決しない場合: ロールバック実施
4. MongoDB 接続エラーの場合: `MONGODB_URI` を確認
