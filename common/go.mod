module github.com/stream-stack/store/store/common

go 1.16

require (
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/envoyproxy/go-control-plane v0.7.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/protobuf v1.3.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/exp v0.0.0-20190121172915-509febef88a4 // indirect
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be // indirect
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	google.golang.org/protobuf v1.26.0-rc.1
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	google.golang.org/grpc v1.27.0
)

replace (
	//github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6
	//google.golang.org/grpc v1.39.0 => google.golang.org/grpc v1.26.0
	github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
	//google.golang.org/grpc => google.golang.org/grpc v1.27.0
	//google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
