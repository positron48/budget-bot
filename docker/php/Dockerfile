FROM php:8.2-fpm

RUN apt-get update && apt-get install -y \
    git \
    unzip \
    libicu-dev \
    libsqlite3-dev \
    && docker-php-ext-install \
    pdo_sqlite \
    intl \
    && pecl install pcov \
    && git clone https://github.com/runkit7/runkit7.git \
    && cd runkit7 \
    && phpize \
    && ./configure \
    && make \
    && make install \
    && docker-php-ext-enable \
    pdo_sqlite \
    pcov \
    runkit7

# Create non-root user
RUN useradd -m -u 1000 appuser \
    && mkdir -p /var/www/budget-bot \
    && chown -R appuser:appuser /var/www/budget-bot

WORKDIR /var/www/budget-bot

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

USER appuser

CMD ["php-fpm"]