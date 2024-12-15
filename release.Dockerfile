FROM scratch
WORKDIR /
COPY ./binary /binary
EXPOSE 8020
CMD ["/binary"]