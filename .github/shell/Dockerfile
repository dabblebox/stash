FROM golang:1.13.7

RUN go get github.com/mitchellh/gox && \
    go get github.com/tcnksm/ghr

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]