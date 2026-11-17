# ---- build stage ----
FROM composer:2 AS build
WORKDIR /app
COPY composer.json ./
RUN composer install --no-dev --no-interaction --optimize-autoloader
COPY . .

# ---- final stage ----
FROM php:8.3-fpm-alpine AS final
ARG GIT_COMMIT=unknown
ENV GIT_COMMIT=$GIT_COMMIT
RUN apk add --no-cache nginx && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /app /app
COPY php-fpm-prod.ini /usr/local/etc/php/conf.d/prod.ini
COPY www.conf /usr/local/etc/php-fpm.d/www.conf
COPY nginx.conf /etc/nginx/http.d/default.conf
COPY start.sh /start.sh
RUN chmod +x /start.sh && \
    mkdir -p /var/lib/nginx/tmp /var/log/nginx /run/nginx && \
    chown -R app:app /app /var/lib/nginx /var/log/nginx /run/nginx /usr/local/etc/php-fpm.d
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -q -O- http://127.0.0.1:8080/health || exit 1
CMD ["/start.sh"]
