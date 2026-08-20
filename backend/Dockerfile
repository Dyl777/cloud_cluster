FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -tags kafka -o bin/identity ./cmd/identity \
 && go build -tags kafka -o bin/wallet ./cmd/wallet \
 && go build -tags kafka -o bin/payments ./cmd/payments \
 && go build -tags kafka -o bin/marketplace ./cmd/marketplace \
 && go build -tags kafka -o bin/provision ./cmd/provision \
 && go build -tags kafka -o bin/settlement ./cmd/settlement \
 && go build -tags kafka -o bin/notify ./cmd/notify \
 && go build -tags kafka -o bin/simbridge ./cmd/simbridge \
 && go build -tags kafka -o bin/paymgr ./cmd/paymgr \
 && go build -tags kafka -o bin/mobilevm ./cmd/mobilevm \
 && go build -tags kafka -o bin/mobilegateway ./cmd/mobilegateway \
 && go build -tags kafka -o bin/compat ./cmd/compat

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /src/bin ./bin
EXPOSE 8081 8082 8083 8084 8085 8086 8088 8089 8090 8091 8092
CMD ["./bin/identity"]
