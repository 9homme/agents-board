# Multi-stage Next.js build.
# Stage 1 — install dependencies and build.
# Stage 2 — run production Next.js server.
#
# Architecture ref: REQ005 §6 (US008 scope), D-012.
# Decision D-012: containerised production build (npm run build + npm start).
# Dev-loop usage: use `startup.sh` or `cd web && npm run dev` directly on the host.

FROM node:20-alpine AS deps

WORKDIR /app

# Install dependencies before copying source (cache-friendly).
COPY package.json package-lock.json ./
RUN npm ci

# Stage 2 — build the Next.js production bundle.
FROM node:20-alpine AS builder

WORKDIR /app

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

CMD ["npm", "start"]
