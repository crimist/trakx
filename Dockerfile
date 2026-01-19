FROM golang:latest

WORKDIR /trakx
COPY . .

RUN go build -v -o trakx ./cmd/trakx

RUN addgroup --system trakx && adduser --system --ingroup trakx --disabled-password trakx 
USER trakx:trakx

EXPOSE 1337/tcp
EXPOSE 1337/udp

CMD ["./trakx", "execute"]
