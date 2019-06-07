FROM golang:1.12-stretch AS lang
MAINTAINER RowBot
WORKDIR /home/source

COPY . .
RUN go get -d && go build -v

FROM ubuntu:18.04

RUN apt-get update && apt-get install -y postgresql-10

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER park_bonus WITH SUPERUSER PASSWORD 'admin';" &&\
    createdb -O park_bonus docker &&\
    # withput this line pgx driver won't work
    psql -d docker -c "CREATE EXTENSION IF NOT EXISTS citext;" &&\
    /etc/init.d/postgresql stop

USER root

RUN echo "listen_addresses = '*'" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "synchronous_commit = off" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "fsync = off" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "full_page_writes = off" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "max_wal_size = 1GB" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "shared_buffers = 512MB" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "effective_cache_size = 256MB" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "work_mem = 64MB" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "maintenance_work_mem = 128MB" >> /etc/postgresql/10/main/postgresql.conf
RUN echo "unix_socket_directories = '/var/run/postgresql'" >> /etc/postgresql/10/main/postgresql.conf

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

EXPOSE 5000

USER postgres

WORKDIR /home/source
COPY --from=lang /home/source .

CMD /etc/init.d/postgresql start && ./source
