FROM golang:alpine as builder

WORKDIR /apigateway

# add git so we can fetch dependencies with go get
RUN apk add --no-cache ca-certificates git

# use go modules for dependencies
COPY go.mod go.sum ./

# fetch dependencies
RUN go mod download

COPY . .

# build go package without CGO
# might need to add -a to force build if the docker cache fails me
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/apigateway .

# create minimal image to save space
FROM scratch

COPY --from=builder /go/bin/apigateway /go/bin/apigateway

ENTRYPOINT ["/go/bin/apigateway"]
