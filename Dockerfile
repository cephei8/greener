############################################################
FROM golang:1.25-trixie AS build

WORKDIR /app

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o greener ./cmd/greener

############################################################
FROM gcr.io/distroless/cc-debian13

WORKDIR /app

COPY --from=build /app/greener /app/greener

EXPOSE 8080
ENTRYPOINT ["/app/greener"]
