FROM nginx:alpine

# Copy frontend HTML and assets
COPY index.html /usr/share/nginx/html/

# Configure NGINX
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose the NGINX port
EXPOSE 80
