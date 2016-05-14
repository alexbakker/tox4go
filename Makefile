all: state-lib state-tool bootstrapd

state-lib: prep
	go build -buildmode=c-shared -o build/lib/libtoxstate.so cmd/state-lib/lib.go

state-tool: prep
	go build -o build/bin/state-tool github.com/Impyy/tox4go/cmd/state-tool

bootstrapd: prep
	go build -o build/bin/bootstrapd github.com/Impyy/tox4go/cmd/bootstrapd

test:
	go test github.com/Impyy/tox4go/crypto github.com/Impyy/tox4go/dht github.com/Impyy/tox4go/state github.com/Impyy/tox4go/transport github.com/Impyy/tox4go/bootstrap

prep:
	mkdir -p build/bin build/lib

clean:
	rm -rf build
