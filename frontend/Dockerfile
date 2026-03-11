# ==========================================
# Base Stage (Shared)
# ==========================================
FROM node:20-alpine AS base
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .

# ==========================================
# Development Stage (Hot Reloading)
# ==========================================
FROM base AS development
EXPOSE 3000
CMD ["npm", "run", "dev"]

# ==========================================
# Build Stage (Compiling Assets)
# ==========================================
FROM base AS builder
RUN npm run build

# ==========================================
# Production Stage (Nginx serving static files)
# ==========================================
FROM nginx:alpine AS production
COPY --from=builder /app/dist /usr/share/nginx/html
# Nginx default config should be enough for static files
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
