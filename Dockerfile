# FROM golang:1.10.3
FROM golang:alpine AS build-env

LABEL maintainer "ericotieno99@gmail.com"
LABEL vendor="Ekas Technologies"

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /go/src/github.com/ekas-data-forwarding

ENV GOOS=linux
ENV GOARCH=386
ENV CGO_ENABLED=0

# Copy the project in to the container
ADD . /go/src/github.com/ekas-data-forwarding

RUN go mod download 

# Go get the project deps
RUN go get github.com/ekas-data-forwarding

# Go install the project
# RUN go install github.com/ekas-data-forwarding
RUN go build

# Set the working environment.
ENV GO_ENV production

# Run the ekas-data-forwarding command by default when the container starts.
# ENTRYPOINT /go/bin/ekas-data-forwarding

# Run the ekas-portal-api command by default when the container starts.
# ENTRYPOINT /go/bin/ekas-portal-api

FROM alpine:latest
WORKDIR /go/

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /go/src/github.com/ekas-data-forwarding/ekas-data-forwarding /go/ekas-data-forwarding

# Use an unprivileged user.
USER appuser

ENTRYPOINT ./ekas-data-forwarding

#Expose the port specific to the ekas API Application.
EXPOSE 6033


