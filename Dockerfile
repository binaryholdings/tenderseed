FROM faddat/archlinux

ENV PATH $PATH:/root/go/bin
ENV GOPATH /root/go/

RUN pacman -Syyu --noconfirm go

RUN go install ./...

CMD tenderseed