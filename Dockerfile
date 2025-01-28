FROM node:alpine3.20

WORKDIR /

COPY bin/app /
COPY config.yml /
COPY package.json /
COPY package-lock.json /
COPY abi.json /
COPY migrations/ /migrations/
COPY internall/ internall/
COPY graph-node/ graph-node/
COPY node_modules/ node_modules/
COPY pkg/ pkg/

RUN npm install

CMD ["./app"]