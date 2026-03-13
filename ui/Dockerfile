# UI Dockerfile for CallSign
# Uses Vite dev server for development (host network mode)
# For production, build a multi-stage image with nginx instead

FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source code
COPY . .

# Expose Vite dev server port
EXPOSE 5173

# Run Vite dev server (host mode for Docker)
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

# ============================================================================
# Production Build (uncomment and use this instead for production deployments)
# ============================================================================
# FROM node:20-alpine AS build
# WORKDIR /app
# COPY package*.json ./
# RUN npm ci
# COPY . .
# RUN npm run build
#
# FROM nginx:alpine
# COPY --from=build /app/dist /usr/share/nginx/html
# EXPOSE 80
# CMD ["nginx", "-g", "daemon off;"]
