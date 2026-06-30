FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:20-alpine3.20 AS builder

WORKDIR /app


RUN apk add --no-cache git ca-certificates && \
    git clone -b master --depth 1 https://gitee.com/dong-xianguo/aiops.git && \
	cd frontend &&\
	npm config set registry https://registry.npmmirror.com && \
	npm ci --registry https://registry.npmmirror.com && \
    npm install && \
	npm run build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/nginx:1.27.4-alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
	apk add --no-cache wget

COPY ./default.conf /etc/nginx/nginx.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 8081
EXPOSE 9200

CMD ["nginx","-g","daemon off;"]
