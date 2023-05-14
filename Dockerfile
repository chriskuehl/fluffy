FROM debian:bullseye

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        dumb-init \
        gcc \
        libxml2-dev \
        libxslt1-dev \
        zlib1g-dev \
        zstd \
    && apt-get clean

RUN curl \
        -sLo /tmp/python.tar.zst \
        'https://github.com/indygreg/python-build-standalone/releases/download/20230116/cpython-3.11.1+20230116-x86_64_v3-unknown-linux-gnu-pgo+lto-full.tar.zst' \
    && echo '8e279b25388e47124a422f300db710cdc98c64cf24bf6903f6f6e8ddbc52d743 */tmp/python.tar.zst' | sha256sum --check \
    && tar -C /opt -xf /tmp/python.tar.zst \
    && rm /tmp/python.tar.zst
ENV PATH=/opt/python/install/bin:$PATH

COPY . /opt/fluffy

RUN install --owner=nobody -d /srv/fluffy
RUN python3.11 -m venv /srv/fluffy/venv \
    && /srv/fluffy/venv/bin/pip install -r /opt/fluffy/requirements.txt /opt/fluffy gunicorn==20.1.0

USER nobody
EXPOSE 8000
ENV FLUFFY_SETTINGS /opt/fluffy/settings/prod_s3.py
ENV PYTHONUNBUFFERED TRUE
CMD [ \
    "/usr/bin/dumb-init", "--", \
    "/srv/fluffy/venv/bin/gunicorn", \
        "-b", "0.0.0.0:8000", \
        "-w", "4", \
        "fluffy.run:app" \
]
