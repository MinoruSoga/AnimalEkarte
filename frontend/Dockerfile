# syntax=docker/dockerfile:1.7
FROM node:24-alpine

RUN apk add --no-cache tzdata && npm install -g pnpm@9

ENV TZ=Asia/Tokyo

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store pnpm install --frozen-lockfile --store-dir /pnpm/store

COPY . .
RUN chmod +x entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["./entrypoint.sh"]
CMD ["pnpm", "dev"]
