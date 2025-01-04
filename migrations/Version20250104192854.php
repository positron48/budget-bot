<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20250104192854 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TEMPORARY TABLE __temp__user AS SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id, state, temp_data FROM user');
        $this->addSql('DROP TABLE user');
        $this->addSql('CREATE TABLE user (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, telegram_id INTEGER NOT NULL, username VARCHAR(255) DEFAULT NULL, first_name VARCHAR(255) DEFAULT NULL, last_name VARCHAR(255) DEFAULT NULL, current_spreadsheet_id VARCHAR(255) DEFAULT NULL, state VARCHAR(255) DEFAULT NULL, temp_data CLOB NOT NULL --(DC2Type:json)
        )');
        $this->addSql('INSERT INTO user (id, telegram_id, username, first_name, last_name, current_spreadsheet_id, state, temp_data) SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id, state, temp_data FROM __temp__user');
        $this->addSql('DROP TABLE __temp__user');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TEMPORARY TABLE __temp__user AS SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id, state, temp_data FROM user');
        $this->addSql('DROP TABLE user');
        $this->addSql('CREATE TABLE user (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, telegram_id INTEGER NOT NULL, username VARCHAR(255) DEFAULT NULL, first_name VARCHAR(255) DEFAULT NULL, last_name VARCHAR(255) DEFAULT NULL, current_spreadsheet_id VARCHAR(255) DEFAULT NULL, state VARCHAR(255) DEFAULT NULL, temp_data CLOB DEFAULT NULL)');
        $this->addSql('INSERT INTO user (id, telegram_id, username, first_name, last_name, current_spreadsheet_id, state, temp_data) SELECT id, telegram_id, username, first_name, last_name, current_spreadsheet_id, state, temp_data FROM __temp__user');
        $this->addSql('DROP TABLE __temp__user');
    }
}
