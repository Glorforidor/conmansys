FROM golang:alpine as builder

WORKDIR /insservice

# add git so we can fetch dependencies with go get
RUN apk update && apk add --update --no-cache ca-certificates git

# use go modules for dependencies
COPY go.mod .
COPY go.sum .

# fetch dependencies
RUN go mod download

COPY . .

# fetch dependencies and build application
# RUN go get -d -v ./... && CGO_ENABLED=0 GOOS=linux go build -a

# build go package without CGO
# might need to add -a to force build if the docker cache fails me
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/insservice .

# create minimal image to save space
FROM scratch

COPY --from=builder /go/bin/insservice /go/bin/insservice

ENTRYPOINT ["/go/bin/insservice"]
