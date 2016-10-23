all: state-lib state-tool bootstrapd dev

dev: monitor

state-lib: prep
	go build -buildmode=c-shared -o build/lib/libtoxstate.so cmd/state-lib/lib.go

state-tool: prep
	go build -o build/bin/state-tool github.com/Impyy/tox4go/cmd/state-tool

bootstrapd: prep
	go build -o build/bin/bootstrapd github.com/Impyy/tox4go/cmd/bootstrapd

monitor: prep
	go run vendor/github.com/Impyy/go-embed/*.go -pkg=main -input=cmd/dev/monitor/assets -output=cmd/dev/monitor/assets.go
	go build -o build/bin/dev/monitor github.com/Impyy/tox4go/cmd/dev/monitor

test:
	go test github.com/Impyy/tox4go/crypto github.com/Impyy/tox4go/dht github.com/Impyy/tox4go/state github.com/Impyy/tox4go/transport github.com/Impyy/tox4go/bootstrap github.com/Impyy/tox4go/relay

prep:
	mkdir -p build/bin build/bin/dev build/lib

clean:
	rm -f cmd/dev/monitor/assets.go
	rm -rf build
