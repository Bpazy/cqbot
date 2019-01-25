FROM golang

RUN cd ~ \
    && git clone https://github.com/Bpazy/cqbot \
    && cd cqbot \
    && go build . \
    && rm -rf /go/* \

ENTRYPOINT ["./cqbot"]