FROM faddat/archlinux AS builder

COPY . /tinyseed


ENV PATH $PATH:/root/go/bin
ENV GOPATH /root/go/

RUN pacman -Syyu --noconfirm go base-devel


RUN cd /tinyseed && go install ./...

FROM faddat/archlinux 

COPY --from=builder /root/go/bin/tinyseed /usr/bin/tinyseed

CMD tinyseed