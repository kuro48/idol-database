#!/bin/bash
# MongoDB Atlas用の環境変数を使用

if [ -f .env.atlas ]; then
    cp .env.atlas .env
    echo "✅ MongoDB Atlas用の設定に切り替えました (.env.atlas → .env)"
    echo "☁️  クラウドのMongoDBに接続します"
else
    echo "❌ .env.atlas ファイルが見つかりません"
    exit 1
fi
