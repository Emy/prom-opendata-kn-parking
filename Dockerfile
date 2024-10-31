FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /prom-opendata-kn-parking

EXPOSE 4276

CMD ["/prom-opendata-kn-parking"]