FROM 499144353299.dkr.ecr.eu-central-1.amazonaws.com/docker-hub/library/golang:1.26-alpine AS build

WORKDIR /build
RUN apk --update add make ca-certificates tzdata git bash


ENV GOPRIVATE=gitlab.smartbet.am
ENV GOSUMDB=off

ARG GIT_MODULE_USER
ARG GIT_MODULE_TOKEN

RUN git config --global url."https://${GIT_MODULE_USER}:${GIT_MODULE_TOKEN}@gitlab.smartbet.am/".insteadOf "https://gitlab.smartbet.am/"

COPY ./go.mod /build/go.mod
COPY ./go.sum /build/go.sum

RUN go mod download

COPY . .

RUN go build -o app

FROM scratch
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /build/app /app
COPY --from=build /build/docs /docs

ENTRYPOINT ["/app"]

