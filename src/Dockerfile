FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server .

FROM alpine:3.19
WORKDIR /app
COPY --from=build /src/server .
EXPOSE 8080
ENTRYPOINT ["./server"]
