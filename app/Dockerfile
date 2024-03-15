FROM golang:1.21.7-bookworm as builder

COPY . /build

WORKDIR /build

RUN set -eux; \
    go build -o /build/go-pg-nats ./cmd


FROM debian:12.2

COPY --from=builder /build/go-pg-nats /app/go-pg-nats
COPY --from=builder /build/server/static /app/static
WORKDIR /app

# Change Timezone to Europe/Moscow
RUN apt-get update && apt-get install tzdata -y
ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone


# Set up user for server
RUN set -eux; \
    chmod +x go-pg-nats; \
    groupadd -r serveruser; \
    useradd -r -g serveruser -d /app -s /bin/bash serveruser; \
    chown -R serveruser:serveruser /app;

USER serveruser

EXPOSE 8080

CMD [ "./go-pg-nats" ]
