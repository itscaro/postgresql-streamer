FROM postgres:11

ENV WAL2JSON_VERSION da90c76a69966a7dfcf0657acacde916164bd9c0

# Postgres + Postgis + sql related deps
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libsybdb5 \
        postgresql-server-dev-11 \
        perl \
        ca-certificates \
        automake autoconf libtool make gcc gettext pkg-config wget build-essential \
        libxml2-dev libgeos-dev libproj-dev libgdal-dev && \
    echo "----> Install wal2json" && \
        wget https://github.com/eulerto/wal2json/archive/${WAL2JSON_VERSION}.tar.gz -O - | tar -xz && \
        cd wal2json-${WAL2JSON_VERSION} && \
        make USE_PGXS=1 && make USE_PGXS=1 install && \
        cd .. && \
        rm -rf wal2json-${WAL2JSON_VERSION} && \
    apt-get autoremove -y --purge \
        libsybdb5 \
        automake autoconf libtool make gcc gettext pkg-config wget build-essential \
        libxml2-dev libgeos-dev libproj-dev libgdal-dev && \
    rm -rf /var/lib/apt/lists/*

COPY postgresql.conf /etc/postgresql/postgresql.conf
COPY pg_hba.conf /etc/postgresql/pg_hba.conf

CMD ["-c", "config_file=/etc/postgresql/postgresql.conf"]
