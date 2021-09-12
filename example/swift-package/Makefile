.DEFAULT: build

build: Frameworks $(wildcard Package.*) $(wildcard Sources/*/*)
	swift build -c release

.PHONY: Frameworks
Frameworks: Frameworks/GetGo.xcframework

Frameworks/%.xcframework: Makefile $(wildcard Sources/*/*.go) $(shell which gomobile)
	mkdir -p Frameworks
	gomobile init
	gomobile bind -target=ios,iossimulator,macos,maccatalyst -x -o $@ ./Sources/$*
	touch $@

test: build
	go test ./...
	swift test
