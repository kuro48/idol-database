#!/bin/bash
# ローカルDocker MongoDB用の環境変数を使用

if [ -f .env.local ]; then
    cp .env.local .env
    echo "✅ ローカルDocker MongoDB用の設定に切り替えました (.env.local → .env)"
    echo "📦 MongoDB起動コマンド: docker-compose up -d"
else
    echo "❌ .env.local ファイルが見つかりません"
    exit 1
fi
