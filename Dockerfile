FROM python:3.14

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        dumb-init \
    && apt-get clean

COPY --from=ghcr.io/astral-sh/uv:0.9.6 /uv /uvx /usr/local/bin/

COPY . /opt/fluffy

RUN cd /opt/fluffy && uv sync --group gunicorn

USER nobody
EXPOSE 8000
WORKDIR /opt/fluffy
ENV FLUFFY_SETTINGS /opt/fluffy/settings/prod_s3.py
ENV PYTHONUNBUFFERED TRUE
CMD [ \
    "/usr/bin/dumb-init", "--", \
    "/opt/fluffy/.venv/bin/gunicorn", \
        "-b", "0.0.0.0:8000", \
        "-w", "4", \
        "fluffy.run:app" \
]
