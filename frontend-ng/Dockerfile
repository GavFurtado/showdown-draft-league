FROM node:22-alpine AS base
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM base AS dev
EXPOSE 4200
CMD ["npx", "ng", "serve", "--host", "0.0.0.0"]

FROM base AS build
RUN npx ng build --configuration production

FROM nginx:alpine AS prod
COPY --from=build /app/dist/frontend-ng/browser /usr/share/nginx/html
EXPOSE 80
