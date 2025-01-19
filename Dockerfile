FROM golang:1.23.3 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY src ./src

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /JustVPN

FROM scratch
COPY --from=builder /JustVPN /JustVPN

EXPOSE 8081

# RUN
CMD [ "/JustVPN" ]
