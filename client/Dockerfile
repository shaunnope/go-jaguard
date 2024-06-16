FROM golang:1.21

WORKDIR /app
COPY . .

RUN go mod download
RUN apt-get update && apt-get install -y protobuf-compiler && apt-get install -y bc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN chmod +x build.sh && ./build.sh

WORKDIR /app/client

RUN chmod +x execute.sh
RUN chmod +x execute_read.sh
RUN chmod +x execute_write.sh
RUN chmod +x execute_1.sh
RUN chmod +x execute_2.sh
RUN chmod +x execute_3.sh
RUN chmod +x execute_4.sh
RUN chmod +x execute_5.sh
RUN chmod +x execute_6.sh
RUN chmod +x execute_7.sh
RUN chmod +x execute_8.sh
RUN chmod +x execute_9.sh
RUN chmod +x execute_10.sh