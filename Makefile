embed:
	statik -src install/ -include "*.html,*.yaml"

install:
	statik -src install/ -include "*.html,*.yaml"
	go install -v -gcflags='-l=4'

build:
	go build -v -gcflags='-l=4' .
