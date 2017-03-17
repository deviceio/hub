FROM centos:7

ENV PATH "$PATH:/usr/local/go/bin"

RUN yum -y install wget

RUN wget http://download.rethinkdb.com/centos/7/x86_64/rethinkdb.repo \
    -O /etc/yum.repos.d/rethinkdb.repo

RUN yum -y install git rethinkdb &&\
    wget https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz -O ~/go.tar.gz &&\
    tar -C /usr/local -xzf ~/go.tar.gz &&\    
    rm ~/go.tar.gz

ADD . /root/go/src/github.com/deviceio/hub

RUN go get -v github.com/deviceio/hub/...
RUN go install github.com/deviceio/hub
RUN /root/go/bin/hub -install

RUN rm -rf /usr/local/go &&\ 
    rm -rf /root/go &&\
    yum -y remove wget git &&\
    yum -y clean all

CMD rethinkdb --bind all & sleep 15 && /opt/deviceio/hub/bin/hub