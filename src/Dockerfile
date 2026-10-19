FROM golang:1.27-alpine AS build
ARG GOARCH=amd64
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -ldflags="-s -w" -o server .

FROM alpine:3.19
WORKDIR /app
COPY --from=build /src/server .
EXPOSE 8080
ENTRYPOINT ["./server"]
