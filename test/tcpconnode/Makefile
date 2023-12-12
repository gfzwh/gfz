export GOPROXY=https://goproxy.cn,direct

default: build

build: export GO111MODULE=on

build:
	rm -rf bin
	go get
	go build  -o tcpconnode main.go

	# mkdir bin
	# cp -rf conf bin/
	# cp -rf tcpconnode bin/
	# cp stop.sh bin/
	# cp start.sh bin/

install:
	cp -rf bin/* ../bin/tcpconnode/

clean:
	rm -rf bin
	rm tcpconnode
