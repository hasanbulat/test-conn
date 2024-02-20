FROM alpine:3

WORKDIR /app

COPY . .

RUN apk add gcompat

CMD sh -c '/app/testconn'