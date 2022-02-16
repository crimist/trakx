# embed generates the statik file for embedding files in trakx
embed:
	statik -src embedded/ -include "*.html,*.yaml" -f -dest embedded/

# setup installs the necessary utilities to embed files in trakx  
setup:
	go get -v github.com/rakyll/statik

# install runs embed, builds, and installs trakx 
install:
	statik -src embedded/ -include "*.html,*.yaml" -f -dest embedded/
	go install -v

# build builds trakx
build:
	go build -v
