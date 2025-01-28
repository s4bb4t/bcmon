FROM node:alpine3.20

WORKDIR /

COPY bin/app /
COPY config.yaml /
COPY migrations/ /migrations/
COPY node_modules/ node_modules/
COPY package.json /
COPY package-lock.json /

RUN npm install

CMD ["./app"]