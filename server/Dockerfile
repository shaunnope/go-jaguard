FROM golang:1.21

WORKDIR /app
COPY . .

RUN go mod download
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN chmod +x build.sh && ./build.sh

WORKDIR /app/server

