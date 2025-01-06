FROM php:8.2-fpm

RUN apt-get update && apt-get install -y \
    git \
    unzip \
    libicu-dev \
    libsqlite3-dev \
    && docker-php-ext-install \
    pdo_sqlite \
    intl \
    && pecl install pcov xdebug \
    && docker-php-ext-enable \
    pdo_sqlite \
    pcov \
    xdebug \
    intl

# Configure Xdebug
RUN echo "xdebug.mode=debug,coverage" >> /usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini \
    && echo "xdebug.client_host=host.docker.internal" >> /usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini \
    && echo "xdebug.client_port=9003" >> /usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini \
    && echo "xdebug.start_with_request=yes" >> /usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini \
    && echo "xdebug.log=/var/log/xdebug.log" >> /usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini \
    && echo "xdebug.idekey=PHPSTORM" >> /usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini

# Create non-root user
RUN useradd -m -u 1000 appuser \
    && mkdir -p /var/www/budget-bot \
    && chown -R appuser:appuser /var/www/budget-bot

WORKDIR /var/www/budget-bot

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

USER appuser

CMD ["php-fpm"]