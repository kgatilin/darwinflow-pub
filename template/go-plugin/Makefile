.PHONY: build clean run install

build:
	@mkdir -p bin
	go build -o bin/myplugin ./cmd/myplugin

clean:
	rm -rf bin/

run: build
	./bin/myplugin

install: build
	cp bin/myplugin /usr/local/bin/
