FROM golang:1.24.2

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o main ./cmd/app

CMD [ "./main" ]
