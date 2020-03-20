FROM golang

ADD . /app
WORKDIR /app
RUN go build
ENTRYPOINT ["./zagent"]
CMD ["--help"]