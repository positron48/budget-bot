<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20250104150145 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE user_spreadsheet (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, user_id INTEGER NOT NULL, spreadsheet_id VARCHAR(255) NOT NULL, title VARCHAR(255) NOT NULL, month VARCHAR(20) NOT NULL, CONSTRAINT FK_EAF1E07DA76ED395 FOREIGN KEY (user_id) REFERENCES user (id) NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('CREATE INDEX IDX_EAF1E07DA76ED395 ON user_spreadsheet (user_id)');
        $this->addSql('CREATE TEMPORARY TABLE __temp__category AS SELECT id, name, type, is_default FROM category');
        $this->addSql('DROP TABLE category');
        $this->addSql('CREATE TABLE category (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name VARCHAR(255) NOT NULL, type VARCHAR(10) NOT NULL, is_default BOOLEAN NOT NULL)');
        $this->addSql('INSERT INTO category (id, name, type, is_default) SELECT id, name, type, is_default FROM __temp__category');
        $this->addSql('DROP TABLE __temp__category');
        $this->addSql('CREATE TEMPORARY TABLE __temp__category_keyword AS SELECT id, category_id, user_category_id, keyword FROM category_keyword');
        $this->addSql('DROP TABLE category_keyword');
        $this->addSql('CREATE TABLE category_keyword (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, category_id INTEGER DEFAULT NULL, user_category_id INTEGER DEFAULT NULL, keyword VARCHAR(255) NOT NULL, CONSTRAINT FK_9D4B314512469DE2 FOREIGN KEY (category_id) REFERENCES category (id) NOT DEFERRABLE INITIALLY IMMEDIATE, CONSTRAINT FK_9D4B3145BB5D5477 FOREIGN KEY (user_category_id) REFERENCES user_category (id) NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('INSERT INTO category_keyword (id, category_id, user_category_id, keyword) SELECT id, category_id, user_category_id, keyword FROM __temp__category_keyword');
        $this->addSql('DROP TABLE __temp__category_keyword');
        $this->addSql('CREATE INDEX IDX_9D4B314512469DE2 ON category_keyword (category_id)');
        $this->addSql('CREATE INDEX IDX_9D4B3145BB5D5477 ON category_keyword (user_category_id)');
        $this->addSql('CREATE TEMPORARY TABLE __temp__user AS SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id FROM user');
        $this->addSql('DROP TABLE user');
        $this->addSql('CREATE TABLE user (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, telegram_id INTEGER NOT NULL, username VARCHAR(255) DEFAULT NULL, first_name VARCHAR(255) DEFAULT NULL, last_name VARCHAR(255) DEFAULT NULL, current_spreadsheet_id VARCHAR(255) DEFAULT NULL)');
        $this->addSql('INSERT INTO user (id, telegram_id, username, first_name, last_name, current_spreadsheet_id) SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id FROM __temp__user');
        $this->addSql('DROP TABLE __temp__user');
        $this->addSql('CREATE TEMPORARY TABLE __temp__user_category AS SELECT id, user_id, name, type FROM user_category');
        $this->addSql('DROP TABLE user_category');
        $this->addSql('CREATE TABLE user_category (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, user_id INTEGER NOT NULL, name VARCHAR(255) NOT NULL, type VARCHAR(10) NOT NULL, CONSTRAINT FK_E6C1FDC1A76ED395 FOREIGN KEY (user_id) REFERENCES user (id) NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('INSERT INTO user_category (id, user_id, name, type) SELECT id, user_id, name, type FROM __temp__user_category');
        $this->addSql('DROP TABLE __temp__user_category');
        $this->addSql('CREATE INDEX IDX_E6C1FDC1A76ED395 ON user_category (user_id)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('DROP TABLE user_spreadsheet');
        $this->addSql('CREATE TEMPORARY TABLE __temp__category AS SELECT id, name, type, is_default FROM category');
        $this->addSql('DROP TABLE category');
        $this->addSql('CREATE TABLE category (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name VARCHAR(255) NOT NULL, type VARCHAR(10) NOT NULL, is_default BOOLEAN DEFAULT 1 NOT NULL)');
        $this->addSql('INSERT INTO category (id, name, type, is_default) SELECT id, name, type, is_default FROM __temp__category');
        $this->addSql('DROP TABLE __temp__category');
        $this->addSql('CREATE TEMPORARY TABLE __temp__category_keyword AS SELECT id, category_id, user_category_id, keyword FROM category_keyword');
        $this->addSql('DROP TABLE category_keyword');
        $this->addSql('CREATE TABLE category_keyword (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, category_id INTEGER DEFAULT NULL, user_category_id INTEGER DEFAULT NULL, keyword VARCHAR(255) NOT NULL, FOREIGN KEY (category_id) REFERENCES category (id) ON UPDATE NO ACTION ON DELETE CASCADE NOT DEFERRABLE INITIALLY IMMEDIATE, FOREIGN KEY (user_category_id) REFERENCES user_category (id) ON UPDATE NO ACTION ON DELETE CASCADE NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('INSERT INTO category_keyword (id, category_id, user_category_id, keyword) SELECT id, category_id, user_category_id, keyword FROM __temp__category_keyword');
        $this->addSql('DROP TABLE __temp__category_keyword');
        $this->addSql('CREATE INDEX IDX_9D4B314512469DE2 ON category_keyword (category_id)');
        $this->addSql('CREATE INDEX IDX_9D4B3145BB5D5477 ON category_keyword (user_category_id)');
        $this->addSql('CREATE TEMPORARY TABLE __temp__user AS SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id FROM user');
        $this->addSql('DROP TABLE user');
        $this->addSql('CREATE TABLE user (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, telegram_id INTEGER NOT NULL, username VARCHAR(255) DEFAULT NULL, first_name VARCHAR(255) DEFAULT NULL, last_name VARCHAR(255) DEFAULT NULL, current_spreadsheet_id VARCHAR(255) DEFAULT NULL)');
        $this->addSql('INSERT INTO user (id, telegram_id, username, first_name, last_name, current_spreadsheet_id) SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id FROM __temp__user');
        $this->addSql('DROP TABLE __temp__user');
        $this->addSql('CREATE UNIQUE INDEX UNIQ_8D93D649F3E6D02F ON user (telegram_id)');
        $this->addSql('CREATE TEMPORARY TABLE __temp__user_category AS SELECT id, user_id, name, type FROM user_category');
        $this->addSql('DROP TABLE user_category');
        $this->addSql('CREATE TABLE user_category (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, user_id INTEGER NOT NULL, name VARCHAR(255) NOT NULL, type VARCHAR(10) NOT NULL, FOREIGN KEY (user_id) REFERENCES user (id) ON UPDATE NO ACTION ON DELETE CASCADE NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('INSERT INTO user_category (id, user_id, name, type) SELECT id, user_id, name, type FROM __temp__user_category');
        $this->addSql('DROP TABLE __temp__user_category');
        $this->addSql('CREATE INDEX IDX_E6C1FDC1A76ED395 ON user_category (user_id)');
    }
}
