FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /fitflow-api ./cmd/api

FROM alpine:3.22

WORKDIR /app
COPY --from=build /fitflow-api /app/fitflow-api

EXPOSE 8080

ENTRYPOINT ["/app/fitflow-api"]
