FROM golang:1.22 as binaryimage

ARG GOPROXY
ARG JUICEFS_REPO_URL=https://github.com/juicedata/juicefs
ARG JUICEFS_REPO_BRANCH=main
ARG JUICEFS_REPO_REF=${JUICEFS_REPO_BRANCH}

COPY . /workspace
WORKDIR /workspace

ENV GOPROXY=${GOPROXY:-https://goproxy.cn}

RUN go install github.com/go-delve/delve/cmd/dlv@v1.22.1

RUN go mod tidy
RUN make juicefs
RUN mv juicefs /usr/local/bin/juicefs

# ----------

FROM golang:1.22
COPY --from=binaryimage /usr/local/bin/juicefs /usr/local/bin/juicefs
COPY --from=binaryimage /go/bin/dlv /usr/local/bin/dlv
WORKDIR /jfs
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources
RUN apt-get update && apt-get install -y wget fuse3 gnupg2 curl htop vim fio less
RUN ln -s /usr/local/bin/juicefs /bin/mount.juicefs && /usr/local/bin/juicefs --version
