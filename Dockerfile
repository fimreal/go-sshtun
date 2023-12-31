FROM golang:latest as builder
COPY . /srv/sshtun
RUN cd /srv/sshtun &&\
    make docker-build &&\
    ls -l bin

# alpine
# FROM alpine:latest
FROM scratch
LABEL source.url="https://github.com/fimreal/go-sshtun"
COPY --from=builder /srv/sshtun/bin/sshtun /sshtun
ENV LISTEN "0.0.0.0:1080"
ENTRYPOINT [ "/sshtun" ]