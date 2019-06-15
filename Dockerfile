FROM golang:1.12-stretch AS lang
WORKDIR /home/db-forums
COPY . .
RUN go get -d && go build -v


FROM ubuntu:18.04
MAINTAINER Andrey
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && apt-get install -y gnupg
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y git

## Клонируем проект
USER root
##RUN git clone https://github.com/rowbotman/db-forums.git
#WORKDIR /home/db-forums
#COPY . .
#RUN go get -d && go build -v
#RUN cd /home/db-forums
WORKDIR /home/db-forums
RUN cd /home/db-forums
COPY . .

# Устанавливаем PostgreSQL
RUN apt-get -y update
RUN apt-get -y install apt-transport-https git wget
RUN echo 'deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main' >> /etc/apt/sources.list.d/pgdg.list
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN apt-get -y update
ENV PGVERSION 11
RUN apt-get -y install postgresql-$PGVERSION postgresql-contrib

USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER park WITH SUPERUSER PASSWORD 'admin';" &&\
    createdb -O park park_forum && psql -d park_forum -c "CREATE EXTENSION IF NOT EXISTS citext;" &&\
    psql park_forum -a -f ./init.sql &&\
    /etc/init.d/postgresql stop
 #fsync = off
 #full_page_writes = off
 #jit = off
 #autovacuum = off
 #synchronous_commit = off
 #archive_mode = off
 #huge_pages = off
 #work_mem = 64MB
 #max_wal_size = 1GB
 #shared_buffers = 512MB
 #effective_cache_size = 256MB
 #maintenance_work_mem = 256MB
 #checkpoint_timeout = 15min
 #unix_socket_directories = '/var/run/postgresql'
 #wal_buffers = 4MB
 #listen_addresses = '*'
USER root
RUN echo "local all all md5" > /etc/postgresql/$PGVERSION/main/pg_hba.conf &&\
    echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVERSION/main/pg_hba.conf &&\
    echo "listen_addresses = '*'" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "fsync = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "synchronous_commit = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "full_page_writes = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "autovacuum = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "unix_socket_directories = '/var/run/postgresql'" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "work_mem = 64MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "huge_pages = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "maintenance_work_mem = 256MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "shared_buffers = 512MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "wal_buffers = 4MB" >> /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "checkpoint_timeout = 15min" >>  /etc/postgresql/$PGVERSION/main/postgresql.conf &&\
    echo "archive_mode = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]
EXPOSE 5432
EXPOSE 5000

WORKDIR /home/db-forums
COPY --from=lang /home/db-forums .

CMD service postgresql start && ./db-forums
