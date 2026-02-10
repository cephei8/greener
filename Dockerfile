############################################################
FROM golang:1.26-trixie AS build

RUN apt-get update && apt-get install -y nodejs npm

WORKDIR /app

COPY . .

RUN npm install
RUN npm run build
RUN CGO_ENABLED=1 GOOS=linux go build -o greener ./cmd/greener

############################################################
FROM gcr.io/distroless/cc-debian13

COPY --from=build /app/greener /app/greener

EXPOSE 8080
ENTRYPOINT ["/app/greener"]
