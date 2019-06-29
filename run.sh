echo "Running Trakx. Git pull and then ^C within the screen to restart the service with new code"
screen -dm bash -c "while true; do go run main.go -x -http=false; done"
screen -list
