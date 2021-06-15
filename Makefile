setup:
	go get -v github.com/rakyll/statik

embed:
	statik -src embeded/ -include "*.html,*.yaml" -f -dest embeded/

install:
	statik -src embeded/ -include "*.html,*.yaml" -f -dest embeded/
	go install -v

build:
	go build -v
