FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o bin/identity ./cmd/identity \
 && go build -o bin/wallet ./cmd/wallet \
 && go build -o bin/payments ./cmd/payments \
 && go build -o bin/marketplace ./cmd/marketplace \
 && go build -o bin/provision ./cmd/provision \
 && go build -o bin/settlement ./cmd/settlement \
 && go build -o bin/notify ./cmd/notify \
 && go build -o bin/simbridge ./cmd/simbridge \
 && go build -o bin/paymgr ./cmd/paymgr \
 && go build -o bin/mobilevm ./cmd/mobilevm \
 && go build -o bin/mobilegateway ./cmd/mobilegateway \
 && go build -o bin/compat ./cmd/compat

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /src/bin ./bin
EXPOSE 8081 8082 8083 8084 8085 8086 8088 8089 8090 8091 8092
CMD ["./bin/identity"]
