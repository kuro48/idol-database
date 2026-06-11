# ---- build stage ----
FROM node:26-alpine AS builder

RUN npm install -g pnpm

WORKDIR /workspace

# swagger.json をビルドコンテキストで参照できるようコピー
COPY backend/docs/swagger.json ./backend/docs/swagger.json

COPY docs/package.json docs/pnpm-lock.yaml* ./docs/
WORKDIR /workspace/docs
RUN pnpm install --frozen-lockfile

COPY docs/ ./
ARG API_HOST
ENV API_HOST=${API_HOST}
RUN pnpm build

# ---- serve stage ----
FROM nginx:1.27-alpine

COPY --from=builder /workspace/docs/dist /usr/share/nginx/html
COPY docs/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
