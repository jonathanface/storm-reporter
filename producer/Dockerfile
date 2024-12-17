FROM node:18

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json, then install dependencies
COPY package*.json ./
RUN npm install

# Copy the rest of the application, including tsconfig.json and src directory
COPY . .

# Compile TypeScript
RUN npx tsc

# Expose the application port (optional)
EXPOSE 3000

# Start the application
CMD ["node", "dist/index.js"]
