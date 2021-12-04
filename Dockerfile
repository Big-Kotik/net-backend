# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /src

COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN mkdir /static
COPY src/*.go ./
COPY static/home.html ./static/

RUN cd /src
RUN go build
CMD [ "./net-backend", "-debug" ]

