FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ue_uml .

FROM scratch

WORKDIR /app

COPY --from=builder /build/ue_uml   .
COPY --from=builder /build/config.yml .

EXPOSE 8080

ENTRYPOINT ["/app/ue_uml"]
