FROM nginx:1.17.3

RUN mkdir -p /etc/ssl/private && \
    rm /etc/nginx/conf.d/default.conf

VOLUME /etc/ssl/private

EXPOSE 80
EXPOSE 443

CMD ["nginx", "-g", "daemon off;"]
