FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang git make && \
    go env -w GOPROXY=https://goproxy.cn,direct

RUN git clone https://github.com/git-lfs/git-lfs.git -b v3.4.0 && \
    cd git-lfs && make

# build binary
COPY . /go/src/github.com/opensourceways/software-package-server
RUN cd /go/src/github.com/opensourceways/software-package-server/message-server && GO111MODULE=on CGO_ENABLED=0 go build

WORKDIR /go/src/github.com/opensourceways/software-package-server/download
RUN cd /go/src/github.com/opensourceways/software-package-server/download && GO111MODULE=on CGO_ENABLED=0 go build -a -o download .

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow curl git && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 message-server && \
    useradd -u 1000 -g message-server -s /sbin/nologin -m message-server && \
    echo "umask 027" >> /home/message-server/.bashrc && \
    echo 'set +o history' >> /home/message-server/.bashrc && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && rm -rf /tmp/* && \
    mkdir /opt/app -p && chmod 700 /opt/app && chown message-server:message-server /opt/app

COPY --chown=root --from=BUILDER /git-lfs/bin/git-lfs /usr/local/bin/git-lfs
COPY --chown=message-server --from=BUILDER /go/src/github.com/opensourceways/software-package-server/message-server/message-server /opt/app/message-server
COPY --chown=message-server --from=BUILDER /go/src/github.com/opensourceways/software-package-server/download/download /opt/app/download
COPY --chown=message-server softwarepkg/infrastructure/pkgciimpl/clone_repo.sh /opt/app/clone_repo.sh
COPY --chown=message-server softwarepkg/infrastructure/pkgciimpl/pull_request.sh /opt/app/pull_request.sh

USER message-server

RUN chmod 550 /opt/app/download /opt/app/message-server /opt/app/clone_repo.sh /opt/app/pull_request.sh

WORKDIR /opt/app/

ENTRYPOINT ["/opt/app/message-server"]
