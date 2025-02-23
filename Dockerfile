ARG GO_VERSION=1.23.5

FROM golang:${GO_VERSION}-alpine AS golang

WORKDIR /app
COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /storage cmd/storage/main.go

FROM gcr.io/distroless/static:latest

COPY --from=golang /storage .

EXPOSE 9500

CMD ["/storage"]
