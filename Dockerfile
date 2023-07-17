FROM golang:latest as builder
COPY . /srv/sshtun
RUN cd /srv/sshtun &&\
    go build -o sshtun &&\
    ls -l

# alpine
# FROM alpine:latest
FROM scratch
LABEL source.url="https://github.com/fimreal/go-sshtun"
COPY --from=builder /srv/sshtun /sshtun
ENTRYPOINT [ "/sshtun" ]