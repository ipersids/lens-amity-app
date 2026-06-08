FROM postgres:18-alpine

# Install pg_cron extention
RUN apk update && \
    apk add --no-cache --virtual .build-deps git make gcc musl-dev clang19 llvm19-dev util-linux-dev && \
    git clone https://github.com/citusdata/pg_cron.git && \
    cd pg_cron && \
    make && make install && \
    cd .. && rm -rf pg_cron && \
    apk del .build-deps
