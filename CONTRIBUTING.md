# Contributing

## Development Setup

```bash
cp .env.example .env.local
docker-compose up -d mongodb
cd backend
go test ./...
go run cmd/api/main.go
```

For frontend work:

```bash
cd frontend
pnpm install --frozen-lockfile
pnpm lint
pnpm build
```

## Pull Request Checklist

- Keep generated secrets and local env files out of Git.
- Run `go test ./...` from `backend/`.
- Run `pnpm lint` and `pnpm build` from `frontend/` when changing frontend code.
- Regenerate Swagger docs when API annotations or request/response contracts change:

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs
```

- Confirm public endpoints, authentication, CORS, and rate limits still match the README.

## Security Expectations

Validate user input at API boundaries, avoid logging tokens or personal data, and prefer environment variables for all secrets. Public write endpoints must stay authenticated and rate limited.
