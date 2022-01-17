# setup installs the necessary utilities to embed files in trakx  
setup:
	go get -v github.com/rakyll/statik

# embed generates the statik file for embedding files in trakx
embed:
	statik -src embeded/ -include "*.html,*.yaml" -f -dest embeded/

# install runs embed, builds, and installs trakx 
install:
	statik -src embeded/ -include "*.html,*.yaml" -f -dest embeded/
	go install -v

# build builds trakx
build:
	go build -v
