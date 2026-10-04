FROM node:20-alpine

WORKDIR /app

# Ensure correct permissions for the non-root 'node' user
RUN chown -R node:node /app

USER node

# Copy manifests first for efficient Docker layer caching
COPY --chown=node:node package*.json ./

# Install all dependencies (including devDependencies)
RUN npm install

# Copy application source code
COPY --chown=node:node . .

EXPOSE 3000

CMD ["npm", "run", "dev"]
