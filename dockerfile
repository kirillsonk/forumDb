FROM ubuntu:18.04

RUN apt-get -y update
ENV PGVER 10
RUN apt-get install -y postgresql-$PGVER
RUN apt install -y golang-1.10 git

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    /etc/init.d/postgresql stop

USER postgres

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

EXPOSE 5000

USER root

ENV GOROOT /usr/lib/go-1.10
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH


USER root

# RUN git clone https://github.com/kirillsonk/forumDb

RUN go get github.com/gorilla/mux
RUN go get github.com/lib/pq
RUN go get github.com/mailru/easyjson...

WORKDIR $GOPATH/src/github.com/kirillsonk/forumDb
ADD . $GOPATH/src/github.com/kirillsonk/forumDb

USER postgres

CMD service postgresql start && go run cmd/main.go