FROM vektra/mockery:v2.8.0 as mockery
FROM bitnami/kubectl:latest as kube
FROM digitalocean/doctl as do

FROM golang:1.15
RUN apt-get update
RUN apt-get -y install apt-transport-https ca-certificates curl gnupg2 software-properties-common
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
RUN apt-get update
RUN apt-get -y install docker-ce
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.33.0
RUN apt-get -y install git

RUN curl -L "https://github.com/docker/compose/releases/download/1.23.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
RUN chmod +x /usr/local/bin/docker-compose

COPY --from=mockery /usr/local/bin/mockery /usr/local/bin/mockery
COPY --from=kube /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl
COPY --from=do /app/doctl /usr/local/bin/doctl

# Forcing go mod and ignore GOPATH
ENV GO111MODULE=on
ENV GO_BIN_FOLDER=/go/bin/

RUN go get github.com/swaggo/swag/cmd/swag@v1.7.0

RUN mkdir -p /go/code

WORKDIR /go/code

ENTRYPOINT ["make"]
