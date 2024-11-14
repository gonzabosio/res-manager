FROM golang:1.23.1 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o view/home
RUN GOARCH=wasm GOOS=js go build -o web/app.wasm

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app .

EXPOSE 3060

CMD ["./view/home", "${BUCKET_NAME}"]