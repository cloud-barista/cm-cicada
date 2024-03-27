FROM python:3.11.8-slim-bookworm
USER root

# MySQL Requirements
RUN apt-get update && apt-get install -y python3-dev default-libmysqlclient-dev build-essential pkg-config

# Python Requirements
ADD requirements.txt .
RUN pip3 install -r requirements.txt
RUN pip3 install connexion
RUN pip3 install "connexion[swagger-ui]"

# Airflow Env Vars
ENV AIRFLOW_HOME='/usr/local/airflow'

# Set wd
WORKDIR /usr/local/airflow

# Sleep forever
CMD sleep infinity