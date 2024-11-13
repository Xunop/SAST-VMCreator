# ----------------------------------------------------------------
# 1. Build the application
# ----------------------------------------------------------------
FROM golang:1.23 AS builder

WORKDIR /app
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vm-creator .

# ----------------------------------------------------------------
# 2. Download depandencies
# ----------------------------------------------------------------
FROM alpine:3.20 AS terraform

ENV TERRAFORM_VERSION=1.9.8

RUN apk add --no-cache wget unzip && \
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    mv terraform /usr/local/bin/terraform && \
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip

# ----------------------------------------------------------------
# 3. Create a minimal image
# ----------------------------------------------------------------
FROM archlinux:base

RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
    bash \
    ca-certificates \
    curl \
    git \
    openssh \
    unzip \
    wget \
    expat \
    openssl \
    xerces-c \
    libxcrypt-compat \
    && \
    pacman -Scc --noconfirm && \
    rm -rf /var/cache/pacman/pkg/*

# Install ovftool binary
COPY --from=builder /app/install-oveftool.sh /app/install-oveftool.sh
RUN /app/install-oveftool.sh

COPY --from=terraform /usr/local/bin/terraform /usr/local/bin/terraform

WORKDIR /app

COPY --from=builder /app/vm-creator /usr/local/bin/vm-creator
COPY --from=builder /app/cloud-init /app/cloud-init
COPY --from=builder /app/terraform /app/terraform

RUN groupadd -r vmgroup && useradd -r -g vmgroup vmuser
RUN chown -R vmuser:vmgroup /app

USER vmuser

# Command to run the executable
ENTRYPOINT ["/usr/local/bin/vm-creator"]
