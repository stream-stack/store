FROM tangxusc/golang:1.18.1 as builder

ENV GOPROXY="https://goproxy.io,direct"
WORKDIR /workspace
COPY * .

# Build
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o store main.go
RUN go build -a -o store cmd/main.go

FROM ubuntu:latest
WORKDIR /
COPY --from=builder /workspace/store .

CMD ["./store"]
