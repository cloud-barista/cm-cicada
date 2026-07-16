# 

# syntax=docker/dockerfile:1.7
FROM python:3.12.5-slim-bookworm AS prod-airflow

# pip/airflow 설정
ENV PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PIP_ROOT_USER_ACTION=ignore \
    AIRFLOW_VERSION=3.3.0 \
    CONSTRAINT_URL=https://raw.githubusercontent.com/apache/airflow/constraints-3.3.0/constraints-3.12.txt \
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

# requirements.txt 는 위 CONSTRAINT_URL(3.3.0) 기준으로 생성한 전체 핀 lock 이다.
# --constraint 를 적용하지 않고 requirements.txt 자체를 단일 기준으로 설치한다.
RUN --mount=type=cache,target=/root/.cache/pip \
    pip3 install --no-cache-dir -r requirements.txt

# Sleep forever
CMD ["sleep", "infinity"]