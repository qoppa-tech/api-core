FROM nginx:latest

RUN touch /var/log/nginx/file_storage.access.log \
  touch /var/log/nginx/file_storage.error.log \
  mkdir /var/lib/nginx

ENTRYPOINT [ "exec", "nginx", "-g", "off;" ]

