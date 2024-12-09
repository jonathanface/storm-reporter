# Step 1: Build the React app using Vite
FROM node:18-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the package.json and lock file to install dependencies
COPY weather-frontend/package*.json ./

# Install dependencies
RUN npm install

# Copy the entire React app source
COPY weather-frontend/ ./

# Build the app for production
RUN npm run build

# Step 2: Use Nginx to serve the built app
FROM nginx:alpine

# Copy the built files from the builder step
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy the Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose the Nginx port
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
