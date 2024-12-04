#!/bin/bash

# Verifica se o arquivo .env já existe, senão cria a partir do .env.example
if [ ! -f /var/www/html/.env ]; then
    cp /var/www/html/.env.example /var/www/html/.env
fi

# Instala as dependências do Laravel usando Composer
composer install --prefer-dist --no-scripts --no-dev

# Instala as dependências do frontend usando npm
npm install --prefer-dist --no-scripts --no-dev

# Executa as migrações, se houver algum banco de dados
php artisan migrate --force
