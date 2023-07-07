FROM alpine
ADD match_helper /match_helper
ENTRYPOINT [ "/match_helper" ]
