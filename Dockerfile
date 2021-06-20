FROM golang:alpine
ADD ./ /project/go-network
WORKDIR /project/go-network
RUN go build -o go-network
ENTRYPOINT ["/project/go-network/go-network"]