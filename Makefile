store-build:
	docker build -f store-Dockerfile -t ccr.ccs.tencentyun.com/k8s-test/test:store-v1 . && docker push ccr.ccs.tencentyun.com/k8s-test/test:store-v1
publisher-build:
	docker build -f publisher-Dockerfile -t ccr.ccs.tencentyun.com/k8s-test/test:publisher-v1 . && docker push ccr.ccs.tencentyun.com/k8s-test/test:publisher-v1
all: store-build publisher-build