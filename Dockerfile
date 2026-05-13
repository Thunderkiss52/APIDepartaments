FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -o /out/app ./cmd/api

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /out/app /app/app
EXPOSE 8080
CMD ["/app/app"]
