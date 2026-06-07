# Multi-stage Next.js build.
# Stage 1 — install dependencies and build.
# Stage 2 — run production Next.js server.
#
# Architecture ref: REQ005 §6 (US008 scope), D-012.
# Decision D-012: containerised production build (npm run build + npm start).
# Dev-loop usage: `make dev-up` (or `cd web && npm run dev`) directly on the host.

FROM node:20-alpine AS deps

WORKDIR /app

# Install dependencies before copying source (cache-friendly).
COPY package.json package-lock.json ./
RUN npm ci

# Stage 2 — build the Next.js production bundle.
FROM node:20-alpine AS builder

WORKDIR /app

# Build args. NEXT_PUBLIC_* env vars are baked into the JS bundle at build time
# by Next.js, NOT read at runtime. Setting them in the compose `environment:` is
# too late — the build must see them. Default points at the e2e stack URL so
# `make e2e` works out of the box.
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npm run build

# Stage 3 — lightweight production runner.
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

# Run as non-root (node:alpine ships with the `node` user, UID 1000).
USER node

CMD ["npm", "start"]
