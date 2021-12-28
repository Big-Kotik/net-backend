# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /src

EXPOSE 8080

COPY go.mod ./
COPY go.sum ./

RUN mkdir /static
COPY src/*.go ./
COPY src/ ./src/

COPY static/home.html ./static/

RUN cd /src
RUN go build
CMD [ "./net-backend", "-debug" ]

