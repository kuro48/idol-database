# MongoDBの公式イメージを使用
FROM mongo:6

# 作業ディレクトリを設定
WORKDIR /data/db

# 認証情報はイメージに焼き込まず、docker-compose.yml または実行環境の
# MONGO_INITDB_ROOT_USERNAME / MONGO_INITDB_ROOT_PASSWORD で渡す。

# 必要に応じて起動スクリプトをコピー
# COPY init.js /docker-entrypoint-initdb.d/

# デフォルトのCMD
CMD ["mongod"]
