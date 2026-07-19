# syntax=docker/dockerfile:1.7
# Vue 3 + Vite SPA — built to static files, served by nginx.
# VITE_* vars are inlined at BUILD time, so the API base URL is a build-arg.
# It defaults to a relative "/api/v1" so the same image works on any host (Caddy
# routes /api/* to the backend, everything else to this container).
ARG NODE_VERSION=22-alpine

# ─────────────── builder ───────────────
FROM node:${NODE_VERSION} AS builder
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

COPY . .
RUN npm run build

# ─────────────── runtime (nginx) ───────────────
FROM nginx:1.27-alpine AS runner
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
