FROM golang:latest
WORKDIR /app
COPY . /app
RUN go mod tidy
RUN go get github.com/prometheus/client_golang/prometheus
RUN go build -o reserva-salas
CMD ["./reserva-salas"]
