# Update
env GIT_TERMINAL_PROMPT=1 go get -u -v github.com/Syc0x00/Trakx

# Setup root if not setup
mkdir -p ~/.trakx/
cp -n config.yaml ~/.trakx/config.yaml
cp -n index.html ~/.trakx/index.html
