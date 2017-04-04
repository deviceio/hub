FROM golang:1.8

ADD . /go/src/github.com/deviceio/hub

RUN cd /go/src/github.com/deviceio/hub && go get -v ./...

RUN go install github.com/deviceio/hub/cmd/deviceio-hub

EXPOSE 443 8975

#USER nobody

CMD ["deviceio-hub", "start"]