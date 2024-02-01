FROM node:20-alpine3.19
RUN apk add --update --no-cache git
WORKDIR /app
RUN npm i @commitlint/cli@18 @commitlint/config-conventional@18
ENTRYPOINT ["npx", "commitlint", "--edit"]
