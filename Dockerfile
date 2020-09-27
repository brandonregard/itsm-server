FROM golang:1.15.2-alpine3.12

ENV GO111MODULE=on
ENV PORT=3000
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build

CMD ["./itsm-server"]