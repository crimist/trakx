FROM golang:latest

WORKDIR /trakx
COPY . .

RUN go build -v

EXPOSE 1337/tcp
EXPOSE 1337/udp

RUN addgroup --system trakx && adduser --system --ingroup trakx --disabled-password trakx 
USER trakx:trakx

CMD ["./trakx", "run"]
