# compile env
FROM golang:1.12

ARG GOARM

ENV CGO_ENABLED 1
ENV GOOS linux
ENV GOARCH arm
ENV CC arm-linux-gnueabihf-gcc
ENV CXX arm-linux-gnueabihf-g++

WORKDIR /app
COPY . /app
RUN go mod download
RUN apt update && apt install -y g++-arm-linux-gnueabihf
RUN go build -o main

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/main /bin/main