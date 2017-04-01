FROM golang:1.8-onbuild

ENV PATH "$PATH:/usr/local/go/bin"

ADD . /root/go/src/github.com/deviceio/hub

RUN yum -y install git wget &&\
    wget https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz -O ~/go.tar.gz &&\
    tar -C /usr/local -xzf ~/go.tar.gz &&\    
    rm ~/go.tar.gz &&\
    go get -v github.com/deviceio/hub/... &&\
    go install github.com/deviceio/hub/cmd/deviceio-hub &&\    
    rm -rf /usr/local/go &&\ 
    rm -rf /root/go &&\
    yum -y remove wget git &&\
    yum -y clean all    

EXPOSE 443 8975

CMD ['hub', 'start']