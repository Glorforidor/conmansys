FROM golang:alpine as builder

WORKDIR /go/src/github.com/Glorforidor/conmansys

# add git so we can fetch dependencies with go get
RUN apk update && apk add git

COPY ./main.go .

# fetch dependencies and build application
RUN go get -d -v ./... && CGO_ENABLED=0 GOOS=linux go build -a

FROM scratch

WORKDIR /app
COPY --from=builder /go/src/github.com/Glorforidor/conmansys/conmansys .

ENTRYPOINT ["./conmansys"]
