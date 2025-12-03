FROM nginx:latest

RUN mkdir --parents /var/lib/nginx/body
RUN touch /var/log/nginx/file_storage.access.log /var/log/nginx/file_storage.error.log

ENTRYPOINT [ "nginx", "-g", "daemon off;" ]