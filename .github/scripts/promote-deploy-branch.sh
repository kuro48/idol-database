#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
deploy_dir="$(mktemp -d)"
payload_dir="$(mktemp -d)"
trap 'rm -rf "$deploy_dir" "$payload_dir"' EXIT

cd "$repo_root"

mkdir -p "$payload_dir/backend/scripts" "$payload_dir/backend/static" "$payload_dir/frontend"

cp docker-compose.yml .env.example .gitignore deploy.sh "$payload_dir"/

cp backend/Dockerfile backend/.dockerignore backend/go.mod backend/go.sum "$payload_dir/backend"/
rsync -a --exclude='*_test.go' backend/cmd "$payload_dir/backend"/
rsync -a --exclude='*_test.go' backend/internal "$payload_dir/backend"/
rsync -a backend/static/terms "$payload_dir/backend/static"/
cp \
  backend/scripts/check-production-env.sh \
  backend/scripts/deploy-production.sh \
  backend/scripts/smoke-production.sh \
  backend/scripts/smoke-read-auth.sh \
  "$payload_dir/backend/scripts"/

cp \
  frontend/package.json \
  frontend/pnpm-lock.yaml \
  frontend/index.html \
  frontend/tsconfig.json \
  frontend/tsconfig.app.json \
  frontend/tsconfig.node.json \
  frontend/vite.config.ts \
  "$payload_dir/frontend"/
rsync -a frontend/public "$payload_dir/frontend"/
rsync -a frontend/src "$payload_dir/frontend"/

cat > "$payload_dir/README.md" <<'EOF'
# idol-api deploy branch

This branch contains only the files needed to deploy the production app.

On the server:

```bash
git checkout deploy
git pull origin deploy
./backend/scripts/deploy-production.sh
```
EOF

if [[ -n "${GITHUB_TOKEN:-}" && -n "${GITHUB_REPOSITORY:-}" ]]; then
  remote_url="https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}.git"
else
  remote_url="$(git -C "$repo_root" remote get-url origin)"
fi

cd "$deploy_dir"
git init
git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
git remote add origin "$remote_url"

if git fetch origin deploy; then
  git checkout -B deploy origin/deploy
else
  git checkout -b deploy
fi

find . -mindepth 1 -maxdepth 1 \
  ! -name '.git' \
  -exec rm -rf {} +
rsync -a "$payload_dir"/ ./

git add -A
if git diff --cached --quiet; then
  echo "deploy branch already up to date"
  exit 0
fi

git commit -m "Promote deploy branch from ${GITHUB_SHA:-unknown}"
git push origin HEAD:deploy
