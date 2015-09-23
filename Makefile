FILENAME = bitrun-api
TARGETS = darwin/amd64 linux/amd64

build:
	go build -o ./bin/$(FILENAME)

all:
	gox \
		-osarch="$(TARGETS)" \
		-output="./bin/$(FILENAME)_{{.OS}}_{{.Arch}}"

setup:
	go get || true

clean:
	rm -rf ./bin/*