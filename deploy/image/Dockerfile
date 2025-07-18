
FROM alpine:latest

WORKDIR /app

COPY alertmanager-hook-adapter  /app/
RUN chmod +x /app/alertmanager-hook-adapter
EXPOSE 8080
ENTRYPOINT ["/app/alertmanager-hook-adapter"]


# docker build --no-cache -t alertmanager-feishu-adapter:v1 -f deploy/image/Dockerfile