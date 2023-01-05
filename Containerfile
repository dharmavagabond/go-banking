# syntax=docker/dockerfile:1.4

###################
####### BASE ######
###################

FROM docker.io/golang:1.19.4-buster AS base

WORKDIR /go/src/go-simple-bank

###################
####### DEV #######
###################

FROM base AS dev

RUN go install github.com/silenceper/gowatch@v1.5.2 && \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh |\
	sh -s -- -b "$(go env GOPATH)"/bin v1.50.1

CMD [ "/go/bin/gowatch" ]

#####################
####### BUILD #######
#####################

FROM base AS build

COPY . .

RUN go mod download && \
	CGO_ENABLED=0 go build -o /go/bin/simple-bank

#####################
####### PROD ########
#####################

FROM gcr.io/distroless/static-debian11 AS prod

COPY --from=build /go/bin/simple-bank /

CMD [ "/simple-bank" ]
