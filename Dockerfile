FROM centos:7

RUN yum -y install wget

RUN wget http://download.rethinkdb.com/centos/7/x86_64/rethinkdb.repo \
          -O /etc/yum.repos.d/rethinkdb.repo

RUN wget https://github.com/deviceio/hub/releases/download/test.v1/deviceio-hub.linux.amd64 \
        -O ~/hub && \
        chmod +x ~/hub

RUN yum -y install rethinkdb

RUN cp /etc/rethinkdb/default.conf.sample /etc/rethinkdb/instances.d/instance1.conf

RUN ~/hub -install

CMD rethinkdb --bind all & sleep 5 && /opt/deviceio/hub/bin/hub