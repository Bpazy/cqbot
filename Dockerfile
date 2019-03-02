FROM golang

RUN cd / \
    && git clone https://github.com/Bpazy/cqbot \
    && cd cqbot \
    && go build -o bot . \
    && rm -rf /go/*

WORKDIR /cqbot

ENTRYPOINT ["./bot"]
