FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang

COPY . /go/src/github.com/opensourceways/software-package-watch
RUN cd /go/src/github.com/opensourceways/software-package-watch/watch && GO111MODULE=on CGO_ENABLED=0 go build

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow curl git && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 watch && \
    useradd -u 1000 -g watch -s /sbin/nologin -m watch && \
    echo "umask 027" >> /home/watch/.bashrc && \
    echo 'set +o history' >> /home/watch/.bashrc && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && rm -rf /tmp/* && \
    mkdir /opt/app -p && chmod 700 /opt/app && chown watch:watch /opt/app

COPY --chown=watch --from=BUILDER /go/src/github.com/opensourceways/software-package-watch/watch/watch /opt/app/watch
COPY --chown=watch softwarepkg/infrastructure/pullrequestimpl/create_branch.sh /opt/app/create_branch.sh
COPY --chown=watch softwarepkg/infrastructure/pullrequestimpl/clone_repo.sh /opt/app/clone_repo.sh
COPY --chown=watch softwarepkg/infrastructure/template /opt/app/template

RUN chmod 550 /opt/app/watch /opt/app/create_branch.sh /opt/app/clone_repo.sh

USER watch

WORKDIR /opt/app/

ENTRYPOINT ["/opt/app/watch"]