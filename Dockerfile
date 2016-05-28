FROM debian:jessie

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        python3 \
        python3-dev \
        virtualenv \
        wget \
    && apt-get clean
RUN wget -O /tmp/dumb-init.deb \
        https://github.com/Yelp/dumb-init/releases/download/v1.0.2/dumb-init_1.0.2_amd64.deb \
    && dpkg -i /tmp/dumb-init.deb \
    && rm /tmp/dumb-init.deb

ADD . /opt/fluffy

RUN install --owner=nobody -d /srv/fluffy
USER nobody
RUN virtualenv -ppython3 /srv/fluffy/venv \
    && /srv/fluffy/venv/bin/pip install /opt/fluffy \
    && /srv/fluffy/venv/bin/pip install gunicorn==19.6

EXPOSE 8000
ENV FLUFFY_SETTINGS /opt/fluffy/settings.py
ENV PYTHONUNBUFFERED TRUE
CMD [ \
    "/usr/bin/dumb-init", "--", \
    "/srv/fluffy/venv/bin/gunicorn", \
        "-b", "0.0.0.0:8000", \
        "-w", "4", \
        "fluffy.run:app" \
]
