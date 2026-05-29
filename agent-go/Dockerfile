FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /hotel-agent ./cmd/checkin-agent

FROM alpine:3.20

WORKDIR /app
COPY --from=build /hotel-agent /app/hotel-agent
RUN mkdir -p /app/logs
CMD ["/app/hotel-agent"]
