# FROM golang:1.23-alpine AS builder
# WORKDIR /
# COPY . .
# RUN apk add --no-cache make && go mod download && make build

FROM node:alpine3.20

WORKDIR /

# COPY --from=builder /bin/app /bin/app

COPY /bin/app /bin/app
COPY package.json /
COPY package-lock.json /
COPY abi.json /
COPY migrations/ migrations/

RUN npm install
# RUN npm install && apk add git

COPY node_modules/ node_modules/

CMD ["/bin/app"]