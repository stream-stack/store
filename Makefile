image-build:
	docker build -f Dockerfile -t ccr.ccs.tencentyun.com/stream/stream:store-v1 .
image-push:
	docker push ccr.ccs.tencentyun.com/stream/stream:store-v1.1
all: image-build image-push
local:
	cd cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o store main.go &&cd .. && docker build -f Dockerfile-local -t ccr.ccs.tencentyun.com/stream/stream:store-v1.1 . && kind load docker-image ccr.ccs.tencentyun.com/stream/stream:store-v1.1 --name c1