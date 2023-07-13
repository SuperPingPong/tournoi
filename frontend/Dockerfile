FROM nginx

COPY ./nginx/default /etc/nginx/conf.d/default.conf
COPY ./html /usr/share/nginx/html
RUN chown -R nginx.nginx /usr/share/nginx/html
