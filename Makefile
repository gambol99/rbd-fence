#
#   Author: Rohith
#   Date: 2015-08-13 16:02:38 +0100 (Thu, 13 Aug 2015)
#
#  vim:ts=2:sw=2:et
#

NAME=rbd-manager
AUTHOR="gambol99"
HARDWARE=$(shell uname -m)
SHA=$(shell git log --pretty=format:'%h' -n 1)
VERSION=$(shell awk '/Version =/ { print $$3 }' cmd/rbd-manager/version.go | sed 's/"//g')

.PHONY: build docker clean release

build:
	mkdir -p ./bin
	mkdir -p ./release
	$(shell sed -i "s/^const GitSha.*/const GitSha = \"${SHA}\"/" cmd/rbd-manager/version.go)
	$(shell cd cmd/rbd-manager && CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o ../../bin/rbd-manager)
	$(shell cd cmd/rbd-unlock && CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o ../../bin/rbd-unlock)

docker: build
	sudo docker build -t ${AUTHOR}/${NAME} .

clean:
	rm -rf ./bin
	rm -rf ./release

test: build
	go get github.com/stretchr/testify
	go test -v ./...

release: build
	mkdir -p release
	$(shell cd cmd/rbd-manager && CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o ../../release/rbd-manager)
	gzip -c release/rbd-manager > release/rbd-manager_${VERSION}_linux_${HARDWARE}.gz
	rm -f release/rbd-manager
