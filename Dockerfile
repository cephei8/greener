############################################################
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build_blazor

RUN apt-get update && apt-get install -y python3

WORKDIR /app
COPY ./client-blazor /app

RUN dotnet workload restore
RUN dotnet restore
RUN dotnet build --no-restore -c Release
RUN dotnet publish GreenerBlazor --no-restore -c Release -o out

############################################################
FROM golang:1-trixie AS build_migration

WORKDIR /app
COPY ./migration /app

RUN go build -o greener-migration

############################################################
FROM python:3.13-slim

RUN apt-get update && apt-get install -y \
    gcc \
    nginx \
    supervisor \
    libpq-dev

RUN mkdir -p /var/log/supervisor

WORKDIR /app
COPY ./server /app

RUN pip install --no-cache-dir -r pip-reqs.txt

COPY --from=build_blazor /app/out/wwwroot /app/blazor
COPY --from=build_migration /app/greener-migration /usr/local/bin/

COPY ./nginx.conf /etc/nginx/nginx.conf
COPY ./supervisord.conf /etc/supervisor/conf.d/supervisord.conf

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 5096

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
