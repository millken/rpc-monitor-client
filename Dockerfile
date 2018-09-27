FROM alpine:3.7

ENV TIMEZONE Asia/Shanghai

LABEL Maintainer="millken <millken@gmail.com>" \
      Description="rpc-monitor-client" 

# persistent / runtime deps
RUN set -ex \
	&& sed -i 's/dl-cdn.alpinelinux.org/mirrors.shu.edu.cn/g' /etc/apk/repositories \
  	&& apk add --update curl bash ca-certificates libc6-compat \
	&& rm -rf /var/cache/apk/*

RUN mkdir -p /data/go
COPY main /data/go

EXPOSE 59051

ENTRYPOINT ["/data/go/main"]