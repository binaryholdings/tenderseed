FROM faddat/archlinux as builder

ENV GOPATH /go
ENV PATH $PATH:/go/bin

RUN pacman -Syyu --noconfirm go git

COPY . ./tenderseed

RUN cd /tenderseed && \ 
      go mod download && \
      go install ./...
      
FROM faddat/archlinux

COPY --from=builder /go/bin/tenderseed /usr/bin/tenderseed

