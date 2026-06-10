FROM node:24-alpine

RUN apk add --no-cache tzdata && npm install -g pnpm@9

ENV TZ=Asia/Tokyo

WORKDIR /app

COPY package.json ./
RUN pnpm install

COPY . .
RUN chmod +x entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["./entrypoint.sh"]
CMD ["pnpm", "dev"]
