FROM alpine:3

RUN apk add curl bash tzdata ca-certificates

ARG UID=1000
ARG GID=1000
ARG USER=app
ARG GROUP=app

RUN addgroup -S ${GROUP} && adduser -S ${USER} -G ${GROUP}


ARG BIN_NAME=oscbridge-6acaf3b4-linux-amd64.bin
COPY --from=build oscbridge-6acaf3b4-linux-amd64.bin /home/$USER/app


# Switch to user
USER ${UID}:${GID}

WORKDIR /home/$USER
ENTRYPOINT ["/bin/bash","-l","-c", "./app"]