FROM golang

ADD . /go/src/github.com/doubtingben/zagent
WORKDIR /go/src/github.com/doubtingben/zagent
RUN go build
ENTRYPOINT ["./zagent"]
CMD ["--help"]