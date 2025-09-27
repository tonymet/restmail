FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o restmail ./cmd/restmail
FROM scratch
WORKDIR /app
COPY --from=builder /app/restmail .
ENTRYPOINT ["./restmail"]
