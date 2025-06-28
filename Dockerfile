# syntax=docker/dockerfile:1.4

###################
####### BASE ######
###################

FROM docker.io/golang:1.24-bullseye AS base

WORKDIR /go/src/go-simple-bank

###################
####### DEV #######
###################

FROM base AS dev

RUN go install github.com/makiuchi-d/arelo@latest && \
	go install github.com/go-task/task/v3/cmd/task@latest && \
	mkdir migrate-cli && cd migrate-cli && \
	curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz && \
	mv migrate /usr/bin && cd .. && rm -rf migrate-cli && \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)"/bin v1.57.2

#####################
####### BUILD #######
#####################

FROM base AS build

COPY . .

RUN go mod download && \
	CGO_ENABLED=0 go build -o /go/bin/sb

#####################
####### PROD ########
#####################

FROM gcr.io/distroless/static-debian11 AS prod

COPY --from=build /go/bin/sb /

CMD [ "/sb" ]
