FROM golang:1.23.3 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY src ./src

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /JustVPN

FROM alpine:latest
WORKDIR /

RUN apk add --no-cache ca-certificates

COPY --from=builder /JustVPN /JustVPN
COPY linode_vpn.tf variables.tf ./secrets.tfvars ./

EXPOSE 8081

# RUN
CMD [ "/JustVPN" ]
