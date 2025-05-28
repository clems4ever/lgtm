FROM golang:1.24

WORKDIR /build

COPY internal/ internal/
COPY main.go main.go
COPY go.sum go.sum
COPY go.mod go.mod

ENV CGO_ENABLED=0

RUN go build -o app main.go

# Create the image serving the website.
FROM alpine
COPY --from=0 /build/app /app