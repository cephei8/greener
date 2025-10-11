FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build_blazor

RUN apt-get update && apt-get install -y python3

WORKDIR /app
COPY ./client-blazor /app

RUN dotnet workload restore
RUN dotnet restore
RUN dotnet build --no-restore -c Release
RUN dotnet publish GreenerBlazor --no-restore -c Release -o out

############################################################

FROM unit:1.34.2-python3.13

WORKDIR /app
COPY ./server /app

RUN pip install --no-cache-dir -r pip-reqs.txt

COPY --from=build_blazor /app/out/wwwroot /app/blazor
COPY ./unit.json /docker-entrypoint.d/

EXPOSE 5096
