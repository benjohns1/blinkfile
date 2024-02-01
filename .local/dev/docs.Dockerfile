FROM hugomods/hugo:0.122.0
COPY go.mod go.sum hugo.yaml /src/
RUN cd /src && hugo mod get
RUN rm -r /src
WORKDIR /src
ENTRYPOINT ["hugo"]