FROM docker.io/library/node:20-alpine AS verify-js
WORKDIR /src
COPY app.js ./app.js
RUN node --check app.js

FROM docker.io/library/nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY . /usr/share/nginx/html
EXPOSE 8081
