FROM fedora:28
MAINTAINER Pacur <contact@pacur.org>

RUN yum -y upgrade
RUN yum -y groupinstall 'Development Tools'
RUN yum -y install tar expect gpg rpm-build rpm-sign openssl createrepo rsync make automake gcc wget zip git bzr mercurial

RUN wget https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
RUN echo "ff54aafedff961eb94792487e827515da683d61a5f9482f668008832631e5d25 go1.11.5.linux-amd64.tar.gz" | sha256sum -c -
RUN tar -C /usr/local -xf go1.11.5.linux-amd64.tar.gz
RUN rm -f go1.11.5.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH /usr/local/go/bin:$PATH:/go/bin

RUN go get github.com/pacur/pacur

ENTRYPOINT ["pacur"]
CMD ["build", "fedora-28"]
