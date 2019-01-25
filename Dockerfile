FROM golang

RUN cd / \
    && git clone https://github.com/Bpazy/cqbot \
    && cd cqbot \
    && go build . \
    && rm -rf /go/*

WORKDIR /cqbot

ENTRYPOINT ["./cqbot"]
