FROM golang:1.22.0-alpine3.19 AS stage_compile
RUN mkdir -p /go/src/sf-hw-26
WORKDIR /go/src/sf-hw-26
ADD main.go .
ADD go.mod .
ADD pipe ./pipe
ADD ring ./ring
RUN go install .
 
FROM alpine:3.19
LABEL version="1.1.0"
WORKDIR /root/
COPY --from=stage_compile /go/bin/sf-hw-26 .
CMD ./sf-hw-26 < ./inout/stdin.txt 1> ./inout/stdout.txt 2> ./inout/stderr.txt