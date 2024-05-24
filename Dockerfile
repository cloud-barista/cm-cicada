FROM alpine:3.19.1

RUN apk --no-cache add tzdata
RUN echo "Asia/Seoul" >  /etc/timezone
RUN cp -f /usr/share/zoneinfo/Asia/Seoul /etc/localtime

RUN mkdir /conf
COPY cmd/cm-cicada/cm-cicada /cm-cicada
RUN mkdir -p /lib/airflow/example/
COPY lib/airflow/example /lib/airflow/example
RUN chmod 755 /cm-cicada

USER root
ENTRYPOINT ["./docker-entrypoint.sh"]
