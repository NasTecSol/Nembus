FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm i --force
COPY . .
RUN npm run build

FROM nginx:alpine
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist/backoffice-frontend/browser /usr/share/nginx/html