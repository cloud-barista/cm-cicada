# 

# syntax=docker/dockerfile:1.7
FROM python:3.12.5-slim-bookworm AS prod-airflow

# pip/airflow 설정
ENV PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PIP_ROOT_USER_ACTION=ignore \
    AIRFLOW_VERSION=2.10.3 \
    CONSTRAINT_URL=https://raw.githubusercontent.com/apache/airflow/constraints-2.10.3/constraints-3.12.txt \
    AIRFLOW_HOME=/usr/local/airflow

RUN --mount=type=cache,target=/var/cache/apt \
    apt-get update && apt-get install -y --no-install-recommends \
        python3-dev \
        build-essential \
        pkg-config \
        default-libmysqlclient-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local/airflow

# Python Requirements
COPY requirements.txt .

# BuildKit 캐시 사용 + constraints 적용
RUN --mount=type=cache,target=/root/.cache/pip \
    pip3 install --no-cache-dir -r requirements.txt --constraint "${CONSTRAINT_URL}"

# Sleep forever
CMD ["sleep", "infinity"]