FROM golang:1.21.0-bullseye 

WORKDIR /cmd/gostarter

ENV CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2 -fstack-protector-all"
ENV GOFLAGS="-buildmode=pie"

COPY go.mod .
COPY main.go .

RUN go get
RUN go build -ldflags "-s -w" -o bin .

USER 65534

ENTRYPOINT [ "/app/bin" ]