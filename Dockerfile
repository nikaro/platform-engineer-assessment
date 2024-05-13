# Build stage
FROM golang:1.22.3-bullseye as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ./
RUN \
    CGO_ENABLED=0 GOOS=linux go build -v -o platform-engineer-assessment && \
    strip platform-engineer-assessment && \
    :

# Final stage
FROM alpine:3.19.1

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=build /app/platform-engineer-assessment ./

USER 1000:1000

ENTRYPOINT ["/app/platform-engineer-assessment"]
