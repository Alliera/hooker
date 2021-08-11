FROM golang:latest
RUN cat /etc/os-release
RUN apt-get update && \
    apt install -y apt-transport-https ca-certificates curl gnupg2 software-properties-common
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
RUN apt update
RUN apt install -y docker-ce docker-ce-cli containerd.io
RUN curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/docker-compose && chmod +x /usr/local/bin/docker-compose && \
        ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose && \
        docker-compose --version
COPY ./main.go /hooker/main.go
COPY ./go.mod /hooker/go.mod
COPY ./docker-compose-internal.yaml /hooker/docker-compose.yaml

RUN cd /hooker && go build
WORKDIR /hooker
ARG GITHUB_TOKEN=token
RUN git clone https://${GITHUB_TOKEN}@github.com/Alliera/xircl.git

CMD service docker start && ./hooker
