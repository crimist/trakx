FROM golang:latest

WORKDIR /trakx

COPY . .

RUN go build -v

EXPOSE 1337

CMD ["./trakx", "run"]
