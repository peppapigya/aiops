FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:20-alpine3.20 AS builder

WORKDIR /app
# 如果是国内的网络建议使用这个镜像仓库，国外的网络不需要
# RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
#     apk add --no-cache git ca-certificates
RUN 

RUN apk add --no-cache git ca-certificates&& \
    git clone -b master --depth 1 https://github.com/peppapigya/aiops.git && \
	cd aiops/frontend &&\
	npm install -g pnpm && \
    pnpm install && \
	pnpm run build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/nginx:1.27.4-alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
	apk add --no-cache wget curl ca-certificates

COPY ./default.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /app/aiops/frontend/dist /usr/share/nginx/html

EXPOSE 8081
EXPOSE 9200

CMD ["nginx","-g","daemon off;"]
