FROM golang:latest
ADD . $GOPATH/src/github.com/Serulian/serulian-langserver/.
WORKDIR $GOPATH/src/github.com/Serulian/serulian-langserver/.
RUN go get -v ./...
WORKDIR $GOPATH/src/github.com/Serulian/serulian-langserver/
RUN go build -o /serulian-langserver .
EXPOSE 4389 
ENTRYPOINT ["/serulian-langserver", "run", "--mode", "tcp"]