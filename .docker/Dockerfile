# MongoDBの公式イメージを使用
FROM mongo:6

# 作業ディレクトリを設定
WORKDIR /data/db

# 環境変数を設定
ENV MONGO_INITDB_ROOT_USERNAME=admin
ENV MONGO_INITDB_ROOT_PASSWORD=password

# 必要に応じて起動スクリプトをコピー
# COPY init.js /docker-entrypoint-initdb.d/

# デフォルトのCMD
CMD ["mongod"]