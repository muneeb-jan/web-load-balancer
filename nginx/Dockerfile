FROM alpine:latest
RUN apk update
RUN apk add nginx
COPY ./nginx.conf /etc/nginx/nginx.conf
RUN mkdir -p /run/nginx
RUN mkdir -p /var/www/nginx
ADD ./www /var/www/nginx/
EXPOSE 80
CMD ["nginx"]
