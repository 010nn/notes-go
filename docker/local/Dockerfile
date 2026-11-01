# --- Stage 1: App Image ---
FROM alpine:latest AS app

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY bin/app .
COPY static/html ./static/html

CMD ["./app"]

# --- Stage 2: Migration Image ---
FROM alpine:latest AS migration

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY bin/migration .
COPY migrations ./migrations

CMD ["./migration"]
