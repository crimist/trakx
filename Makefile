setup:
	go get -v github.com/rakyll/statik

embed:
	statik -src embeded/ -include "*.html,*.yaml" -f -dest embeded/

install:
	statik -src embeded/ -include "*.html,*.yaml" -f -dest embeded/
	go install -v -gcflags='-l=4'

build:
	go build -v -gcflags='-l=4' .
