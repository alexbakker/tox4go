all: state-lib state-tool bootstrapd dev

dev: monitor

state-lib: prep
	go build -buildmode=c-shared -o build/lib/libtoxstate.so cmd/state-lib/lib.go

state-tool: prep
	go build -o build/bin/state-tool github.com/alexbakker/tox4go/cmd/state-tool

bootstrapd: prep
	go build -o build/bin/bootstrapd github.com/alexbakker/tox4go/cmd/bootstrapd

monitor: prep
	go run vendor/github.com/alexbakker/go-embed/*.go -pkg=main -input=cmd/dev/monitor/assets -output=cmd/dev/monitor/assets.go
	go build -o build/bin/dev/monitor github.com/alexbakker/tox4go/cmd/dev/monitor

test:
	go test github.com/alexbakker/tox4go/crypto github.com/alexbakker/tox4go/dht github.com/alexbakker/tox4go/state github.com/alexbakker/tox4go/transport github.com/alexbakker/tox4go/bootstrap github.com/alexbakker/tox4go/relay

prep:
	mkdir -p build/bin build/bin/dev build/lib

clean:
	rm -f cmd/dev/monitor/assets.go
	rm -rf build
