FROM 499144353299.dkr.ecr.eu-central-1.amazonaws.com/docker-hub/library/golang:1.24.1 AS builder

# Set Go env
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/smart-image

# Install dependencies
#RUN apt --update --no-cache add ca-certificates gcc libtool make musl-dev protoc git
RUN apt-get update && apt-get install -y \
    ca-certificates \
    gcc \
    libtool \
    make \
    protobuf-compiler \
    git && \
    rm -rf /var/lib/apt/lists/*


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
