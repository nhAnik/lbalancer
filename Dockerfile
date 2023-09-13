FROM golang:1.21-alpine

WORKDIR /workspace/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build

CMD [ "./lbalancer" ]
