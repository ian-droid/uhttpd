FROM golang:1.10.4-alpine3.8

RUN apk add --no-cache \
 git ;\
 go get github.com/ian-droid/uhttpd

EXPOSE 80 443

VOLUME /www /log /server.crt /server.key

ENTRYPOINT ["uhttpd", "-dir=/www/", "-log=true", "-log-prefix=rhttpd", "-log-path=/log/access.log", "-tls=true", "-cert=/server.crt", "-key=/server.key"]
