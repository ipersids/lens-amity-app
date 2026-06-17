FROM node:24

WORKDIR /usr/src/app

RUN corepack enable

COPY . .
RUN pnpm install --frozen-lockfile

CMD ["pnpm", "run", "dev", "--", "--host", "0.0.0.0"]
