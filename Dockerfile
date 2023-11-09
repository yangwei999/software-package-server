FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang git make && \
    go env -w GOPROXY=https://goproxy.cn,direct

# build binary
COPY . /go/src/github.com/opensourceways/software-package-server
RUN cd /go/src/github.com/opensourceways/software-package-server && GO111MODULE=on CGO_ENABLED=0 go build

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 software-package-server && \
    useradd -u 1000 -g software-package-server -s /sbin/nologin -m software-package-server && \
    echo "umask 027" >> /home/software-package-server/.bashrc && \
    echo 'set +o history' >> /home/software-package-server/.bashrc && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && rm -rf /tmp/* && \
    mkdir /opt/app -p && chmod 700 /opt/app && chown software-package-server:software-package-server /opt/app

USER software-package-server

WORKDIR /opt/app/

COPY --chown=software-package-server --from=BUILDER /go/src/github.com/opensourceways/software-package-server/software-package-server /opt/app

RUN chmod 550 /opt/app/software-package-server

ENTRYPOINT ["/opt/app/software-package-server"]
