FROM golang:1.18
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o /opt/app
ENTRYPOINT ["/opt/app"]