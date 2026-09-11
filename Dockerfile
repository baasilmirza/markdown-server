FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/markdown-server .

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/markdown-server /usr/local/bin/markdown-server
USER app
WORKDIR /docs
EXPOSE 8080
ENTRYPOINT ["markdown-server"]
CMD ["-dir", "/docs", "-addr", ":8080"]
