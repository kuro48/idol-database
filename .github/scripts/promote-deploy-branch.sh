#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
deploy_dir="$(mktemp -d)"
trap 'rm -rf "$deploy_dir"' EXIT

cd "$repo_root"

mkdir -p "$deploy_dir/backend/scripts" "$deploy_dir/backend/static" "$deploy_dir/frontend"

cp docker-compose.yml .env.example .gitignore "$deploy_dir"/

cp backend/Dockerfile backend/.dockerignore backend/go.mod backend/go.sum "$deploy_dir/backend"/
rsync -a --exclude='*_test.go' backend/cmd "$deploy_dir/backend"/
rsync -a --exclude='*_test.go' backend/internal "$deploy_dir/backend"/
rsync -a backend/static/terms "$deploy_dir/backend/static"/
cp \
  backend/scripts/check-production-env.sh \
  backend/scripts/deploy-production.sh \
  backend/scripts/smoke-production.sh \
  backend/scripts/smoke-read-auth.sh \
  "$deploy_dir/backend/scripts"/

cp \
  frontend/package.json \
  frontend/pnpm-lock.yaml \
  frontend/index.html \
  frontend/tsconfig.json \
  frontend/tsconfig.app.json \
  frontend/tsconfig.node.json \
  frontend/vite.config.ts \
  "$deploy_dir/frontend"/
rsync -a frontend/public "$deploy_dir/frontend"/
rsync -a frontend/src "$deploy_dir/frontend"/

cat > "$deploy_dir/README.md" <<'EOF'
# idol-api deploy branch

This branch contains only the files needed to deploy the production app.

On the server:

```bash
git checkout deploy
git pull --ff-only origin deploy
./backend/scripts/deploy-production.sh
```
EOF

cd "$deploy_dir"
git init
git checkout -b deploy
git add .
git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
git commit -m "Promote deploy branch from ${GITHUB_SHA:-unknown}"
if [[ -n "${GITHUB_TOKEN:-}" && -n "${GITHUB_REPOSITORY:-}" ]]; then
  git remote add origin "https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}.git"
else
  git remote add origin "$(git -C "$repo_root" remote get-url origin)"
fi
git fetch origin deploy:refs/remotes/origin/deploy || true
git push origin HEAD:deploy --force-with-lease
