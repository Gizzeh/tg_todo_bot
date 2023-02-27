FROM golang:1.18-alpine3.15 as builder

WORKDIR /project

COPY app/go.mod .
COPY app/go.sum .
RUN go mod download

COPY app/ .

RUN go build

FROM alpine:3.14

COPY --from=builder /project/tg_todo_bot /tg_todo_bot
COPY --from=builder /project/migrations /migrations

EXPOSE 8085