<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20231212000001 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Create category tables';
    }

    public function up(Schema $schema): void
    {
        // Create default categories table
        $this->addSql('CREATE TABLE category (
            id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            name VARCHAR(255) NOT NULL,
            type VARCHAR(10) NOT NULL,
            is_default BOOLEAN NOT NULL DEFAULT 1
        )');

        // Create user categories table
        $this->addSql('CREATE TABLE user_category (
            id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            user_id INTEGER NOT NULL,
            name VARCHAR(255) NOT NULL,
            type VARCHAR(10) NOT NULL,
            FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE
        )');

        // Create category keywords table
        $this->addSql('CREATE TABLE category_keyword (
            id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            category_id INTEGER NULL,
            user_category_id INTEGER NULL,
            keyword VARCHAR(255) NOT NULL,
            FOREIGN KEY(category_id) REFERENCES category(id) ON DELETE CASCADE,
            FOREIGN KEY(user_category_id) REFERENCES user_category(id) ON DELETE CASCADE
        )');

        // Insert default expense categories
        $expenseCategories = [
            'Питание',
            'Подарки',
            'Здоровье/медицина',
            'Дом',
            'Транспорт',
            'Личные расходы',
            'Домашние животные',
            'Коммунальные услуги',
            'Путешествия',
            'Одежда',
            'Развлечения',
            'Кафе/Ресторан',
            'Алко',
            'Образование',
            'Услуги',
            'Авто'
        ];

        foreach ($expenseCategories as $category) {
            $this->addSql("INSERT INTO category (name, type) VALUES (?, 'expense')", [$category]);
        }

        // Insert default income categories
        $incomeCategories = [
            'Зарплата',
            'Премия',
            'Кешбек, др. бонусы',
            'Процентный доход',
            'Инвестиции',
            'Другое'
        ];

        foreach ($incomeCategories as $category) {
            $this->addSql("INSERT INTO category (name, type) VALUES (?, 'income')", [$category]);
        }
    }

    public function down(Schema $schema): void
    {
        $this->addSql('DROP TABLE category_keyword');
        $this->addSql('DROP TABLE user_category');
        $this->addSql('DROP TABLE category');
    }
} 