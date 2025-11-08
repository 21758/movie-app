FROM golang:1.23-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
ENV GOPRIVATE=
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN apk --no-cache add mysql-client

WORKDIR /root/

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

RUN chown -R app:app /root/

USER app

EXPOSE 8080

CMD ["./main"]