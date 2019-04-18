FROM ubuntu:latest

WORKDIR /app
ADD ./conmansys .

ENTRYPOINT ["./conmansys"]
