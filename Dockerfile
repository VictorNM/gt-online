FROM golang:1.17 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./app

# final stage
FROM alpine:latest
# Copy binary from builder
COPY --from=builder /app/app /app
COPY --from=builder /app/config /config
ENV TZ Asia/Ho_Chi_Minh
ENTRYPOINT ["/app"]
EXPOSE 8080