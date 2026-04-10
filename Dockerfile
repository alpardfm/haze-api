FROM golang:1.24.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/migrate ./cmd/migrate \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/seed-admin ./cmd/seed-admin \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/reminder-worker ./cmd/reminder-worker \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/status-worker ./cmd/status-worker

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /out/seed-admin /app/seed-admin
COPY --from=builder /out/reminder-worker /app/reminder-worker
COPY --from=builder /out/status-worker /app/status-worker
COPY migrations /app/migrations
COPY docs /app/docs

EXPOSE 8080

CMD ["/app/api"]

