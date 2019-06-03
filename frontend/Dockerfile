FROM golang:alpine as builder

WORKDIR /frontend

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
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/frontend .

# create minimal image to save space
FROM scratch

WORKDIR /go/bin/

COPY --from=builder /frontend/static/ ./static
COPY --from=builder /frontend/templates/ ./templates
COPY --from=builder /go/bin/frontend .

ENTRYPOINT ["/go/bin/frontend"]