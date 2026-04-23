FROM node:24-alpine

RUN npm install -g pnpm@9

WORKDIR /app

COPY package.json ./
RUN pnpm install

COPY . .
RUN chmod +x entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["./entrypoint.sh"]
CMD ["pnpm", "dev"]
