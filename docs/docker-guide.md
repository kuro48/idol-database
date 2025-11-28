# Docker環境ガイド

このプロジェクトでは、ローカルDockerのMongoDBとMongoDB Atlas（クラウド）の両方を使い分けることができます。

## 📋 環境の種類

### 1. **ローカルDocker MongoDB**
- 開発・テスト用
- インターネット不要
- データはローカルに保存
- 素早く起動・停止可能

### 2. **MongoDB Atlas（クラウド）**
- 本番環境や共有開発用
- インターネット接続が必要
- データはクラウドに保存
- チームで共有可能

---

## 🚀 使い方

### **ローカルDocker MongoDBを使う場合**

#### ステップ1: 環境を切り替え
```bash
# ローカル用の設定に切り替え
bash scripts/use-local.sh
```

#### ステップ2: MongoDBコンテナを起動
```bash
# MongoDBコンテナを起動
docker-compose up -d

# 起動確認
docker-compose ps
```

#### ステップ3: APIサーバーを起動
```bash
# Goアプリケーションを起動
go run cmd/api/main.go
```

#### ステップ4: 動作確認
```bash
# ヘルスチェック
curl http://localhost:8081/

# 期待される出力: {"message":"Hello World"}
```

#### ステップ5: 使用後の停止
```bash
# MongoDBコンテナを停止
docker-compose down

# データも削除する場合（注意！）
docker-compose down -v
```

---

### **MongoDB Atlasを使う場合**

#### ステップ1: 環境を切り替え
```bash
# Atlas用の設定に切り替え
bash scripts/use-atlas.sh
```

#### ステップ2: APIサーバーを起動
```bash
# Goアプリケーションを起動（MongoDBはクラウドなのでdocker-composeは不要）
go run cmd/api/main.go
```

#### ステップ3: 動作確認
```bash
# ヘルスチェック
curl http://localhost:8081/

# 期待される出力: {"message":"Hello World"}
```

---

## 📁 ファイル構成

```
idol-api/
├── .env                 # 現在使用中の環境変数（自動生成、Git管理外）
├── .env.local           # ローカルDocker用の設定
├── .env.atlas           # MongoDB Atlas用の設定
├── .env.example         # サンプル設定ファイル
├── docker-compose.yml   # ローカルMongoDB用のDocker設定
├── .docker/
│   └── Dockerfile       # MongoDBコンテナの設定
└── scripts/
    ├── use-local.sh     # ローカル環境切り替えスクリプト
    └── use-atlas.sh     # Atlas環境切り替えスクリプト
```

---

## 🔧 環境変数の詳細

### **.env.local（ローカルDocker用）**
```bash
MONGODB_URI=mongodb://admin:password@localhost:27017
MONGODB_DATABASE=idol_database
SERVER_PORT=8081
GIN_MODE=debug
```

### **.env.atlas（MongoDB Atlas用）**
```bash
MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/
MONGODB_DATABASE=idol_database
SERVER_PORT=8081
GIN_MODE=debug
```

---

## 💡 よくある使い方

### **パターン1: 通常の開発**
```bash
# ローカルDockerで開発
bash scripts/use-local.sh
docker-compose up -d
go run cmd/api/main.go
```

### **パターン2: チームとデータ共有**
```bash
# Atlasで共有環境を使用
bash scripts/use-atlas.sh
go run cmd/api/main.go
```

### **パターン3: 環境切り替え**
```bash
# ローカル → Atlas
docker-compose down           # ローカルMongoDB停止
bash scripts/use-atlas.sh     # Atlas設定に切り替え
go run cmd/api/main.go        # 再起動

# Atlas → ローカル
# （APIサーバーを停止）
bash scripts/use-local.sh     # ローカル設定に切り替え
docker-compose up -d          # ローカルMongoDB起動
go run cmd/api/main.go        # 再起動
```

---

## 🐛 トラブルシューティング

### 問題1: `docker-compose up -d`が失敗する
**原因**: Dockerが起動していない
**解決策**:
```bash
# Docker Desktopを起動してから再実行
docker-compose up -d
```

### 問題2: MongoDBに接続できない（ローカル）
**原因**: MongoDBコンテナが起動していない
**解決策**:
```bash
# コンテナの状態確認
docker-compose ps

# 再起動
docker-compose restart mongodb
```

### 問題3: MongoDBに接続できない（Atlas）
**原因**: 接続文字列が間違っている、またはネットワークエラー
**解決策**:
```bash
# .env.atlasの接続文字列を確認
cat .env.atlas

# MongoDB Atlasのダッシュボードで接続文字列を再確認
# https://cloud.mongodb.com/
```

### 問題4: ポート27017が既に使われている
**原因**: 別のMongoDBプロセスが実行中
**解決策**:
```bash
# 既存のMongoDBプロセスを確認
lsof -i :27017

# 必要に応じてプロセスを終了
kill -9 <PID>
```

### 問題5: データが消えた
**原因**: `docker-compose down -v`でボリュームを削除した
**解決策**:
- **ローカル**: データはボリュームに保存されるため、`-v`オプションなしで停止すること
- **Atlas**: クラウドなので心配不要

---

## ⚙️ Docker Composeコマンド早見表

| コマンド | 説明 |
|---------|------|
| `docker-compose up -d` | バックグラウンドで起動 |
| `docker-compose ps` | コンテナの状態確認 |
| `docker-compose logs mongodb` | MongoDBのログ確認 |
| `docker-compose restart mongodb` | MongoDB再起動 |
| `docker-compose down` | コンテナ停止（データは保持） |
| `docker-compose down -v` | コンテナ停止 + データ削除 ⚠️ |

---

## 🔐 セキュリティ注意事項

### ⚠️ **重要**: 本番環境の設定

現在のDocker設定は**開発環境専用**です。本番環境では以下を変更してください：

1. **パスワードを強固に**:
```dockerfile
# .docker/Dockerfile
ENV MONGO_INITDB_ROOT_USERNAME=admin
ENV MONGO_INITDB_ROOT_PASSWORD=your_strong_password_here
```

2. **環境変数を外部化**:
```yaml
# docker-compose.yml
environment:
  MONGO_INITDB_ROOT_USERNAME: ${MONGO_USERNAME}
  MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWORD}
```

3. **認証を有効化**:
```bash
# MongoDBコンテナ内で認証を強制
--auth フラグを追加
```

---

## 📊 データの保存場所

### **ローカルDocker**
- **場所**: Dockerボリューム `mongodb_data`
- **確認方法**:
```bash
docker volume ls
docker volume inspect mongodb_data
```

### **MongoDB Atlas**
- **場所**: クラウド（MongoDB Atlasのクラスター）
- **確認方法**: https://cloud.mongodb.com/ のダッシュボード

---

## 🎯 おすすめの開発フロー

### **日常の開発**
1. ローカルDockerで開発・テスト
2. コミット前に動作確認
3. 問題なければGitにプッシュ

### **チーム共有が必要な時**
1. Atlas環境に切り替え
2. チームメンバーと同じデータで確認
3. 完了したらローカルに戻す

### **本番デプロイ前**
1. Atlas環境で最終確認
2. 本番環境の設定を準備
3. デプロイ

---

**最終更新**: 2025-11-14
