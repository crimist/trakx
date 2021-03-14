setup:
	go get -v github.com/rakyll/statik

embed:
	statik -src install/ -include "*.html,*.yaml" -f

install:
	statik -src install/ -include "*.html,*.yaml" -f
	go install -v -gcflags='-l=4'

build:
	go build -v -gcflags='-l=4' .
