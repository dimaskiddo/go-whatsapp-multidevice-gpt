# Builder Image
# ---------------------------------------------------
FROM dimaskiddo/alpine:go-1.19 AS go-builder

WORKDIR /usr/src/app

COPY . ./

RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -o main cmd/main/main.go


# Final Image
# ---------------------------------------------------
FROM dimaskiddo/alpine:base
MAINTAINER Dimas Restu Hidayanto <dimas.restu@student.upi.edu>

ARG SERVICE_NAME="go-whatsapp-multidevice-gpt"

ENV PATH $PATH:/usr/app/${SERVICE_NAME}

WORKDIR /usr/app/${SERVICE_NAME}

RUN mkdir -p dbs \
    && chmod 775 dbs

COPY --from=go-builder /usr/src/app/.env.example ./.env
COPY --from=go-builder /usr/src/app/main ./go-whatsapp-multidevice-gpt

VOLUME ["/usr/app/${SERVICE_NAME}/dbs"]
CMD ["go-whatsapp-multidevice-gpt", "daemon"]
