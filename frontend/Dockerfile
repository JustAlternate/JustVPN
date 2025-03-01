FROM nginx:alpine as builder

# Copy the frontend files to a temporary directory
COPY . /tmp/build

# Set build arguments with defaults
ARG API_BASE_URL=http://localhost:8081

# Replace the API_BASE_URL placeholder in config.js
RUN sed -i "s|\${API_BASE_URL}|${API_BASE_URL}|g" /tmp/build/config.js

# Final stage
FROM nginx:alpine

# Copy the built files from the builder stage
COPY --from=builder /tmp/build /usr/share/nginx/html

# Copy a custom nginx configuration
RUN echo 'server { \
    listen 80; \
    server_name localhost; \
    root /usr/share/nginx/html; \
    index index.html; \
    location / { \
        try_files $uri $uri/ /index.html; \
    } \
    location /api/ { \
        proxy_pass http://justvpn:8081/; \
        proxy_http_version 1.1; \
        proxy_set_header Upgrade $http_upgrade; \
        proxy_set_header Connection "upgrade"; \
        proxy_set_header Host $host; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
