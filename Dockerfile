FROM alpine
ADD html /html
ADD post_api-web /post_api-web
WORKDIR /
ENTRYPOINT [ "/post_api-web" ]
