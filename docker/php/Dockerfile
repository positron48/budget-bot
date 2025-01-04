FROM php:8.2-fpm

RUN apt-get update && apt-get install -y \
    git \
    unzip \
    libicu-dev \
    libsqlite3-dev \
    && docker-php-ext-install \
    pdo_sqlite \
    intl \
    && docker-php-ext-enable \
    pdo_sqlite

WORKDIR /var/www/budget-bot

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

CMD ["php-fpm"] 