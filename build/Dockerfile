FROM alpine:3.8

ENV OPERATOR=/usr/local/bin/business-app-reconciler-controller \
    USER_UID=1001 \
    USER_NAME=business-app-reconciler-controller

# install operator binary
COPY business-app-reconciler-controller ${OPERATOR}

COPY bin /usr/local/bin
RUN  chmod u+x /usr/local/bin/user_setup && chmod ugo+x /usr/local/bin/entrypoint && /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
