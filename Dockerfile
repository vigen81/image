FROM 499144353299.dkr.ecr.eu-central-1.amazonaws.com/docker-hub/library/golang:1.26 AS builder

# Set Go env
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/smart-image

# Install dependencies
#RUN apt --update --no-cache add ca-certificates gcc libtool make musl-dev protoc git
RUN apk --update add make ca-certificates tzdata git bash


ENV GOPRIVATE=gitlab.smartbet.am
ENV GOSUMDB=off

ARG GIT_MODULE_USER
ARG GIT_MODULE_TOKEN

RUN git config --global url."https://${GIT_MODULE_USER}:${GIT_MODULE_TOKEN}@gitlab.smartbet.am/".insteadOf "https://gitlab.smartbet.am/"



RUN git config --global url."https://${GIT_MODULE_USER}:${GIT_MODULE_TOKEN}@gitlab.smartbet.am/".insteadOf "https://gitlab.smartbet.am/"


# Build Go binary
COPY Makefile go.mod go.sum ./
RUN go mod download
COPY . .
RUN make deps
RUN make build

# Deployment container
FROM scratch

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /go/src/smart-image/app /app
ENTRYPOINT ["/app"]
