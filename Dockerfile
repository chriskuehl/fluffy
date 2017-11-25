FROM debian:stretch

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        dumb-init \
        gcc \
        libxml2-dev \
        libxslt1-dev \
        python3 \
        python3-dev \
        virtualenv \
        wget \
        zlib1g-dev \
    && apt-get clean

COPY . /opt/fluffy

RUN install --owner=nobody -d /srv/fluffy
USER nobody
RUN virtualenv -ppython3 /srv/fluffy/venv \
    && /srv/fluffy/venv/bin/pip install /opt/fluffy \
    && /srv/fluffy/venv/bin/pip install gunicorn==19.6

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
