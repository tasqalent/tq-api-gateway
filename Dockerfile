FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gateway ./cmd/gateway

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=build /out/gateway /app/gateway

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/gateway"]