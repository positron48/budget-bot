<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20250104175914 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TEMPORARY TABLE __temp__user_spreadsheet AS SELECT id, user_id, spreadsheet_id, title, month, year FROM user_spreadsheet');
        $this->addSql('DROP TABLE user_spreadsheet');
        $this->addSql('CREATE TABLE user_spreadsheet (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, user_id INTEGER NOT NULL, spreadsheet_id VARCHAR(255) NOT NULL, title VARCHAR(255) NOT NULL, month INTEGER NOT NULL, year INTEGER NOT NULL, CONSTRAINT FK_EAF1E07DA76ED395 FOREIGN KEY (user_id) REFERENCES user (id) ON UPDATE NO ACTION ON DELETE NO ACTION NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('INSERT INTO user_spreadsheet (id, user_id, spreadsheet_id, title, month, year) SELECT id, user_id, spreadsheet_id, title, month, year FROM __temp__user_spreadsheet');
        $this->addSql('DROP TABLE __temp__user_spreadsheet');
        $this->addSql('CREATE INDEX IDX_EAF1E07DA76ED395 ON user_spreadsheet (user_id)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TEMPORARY TABLE __temp__user_spreadsheet AS SELECT id, user_id, spreadsheet_id, title, month, year FROM user_spreadsheet');
        $this->addSql('DROP TABLE user_spreadsheet');
        $this->addSql('CREATE TABLE user_spreadsheet (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, user_id INTEGER NOT NULL, spreadsheet_id VARCHAR(255) NOT NULL, title VARCHAR(255) NOT NULL, month VARCHAR(20) NOT NULL, year INTEGER NOT NULL, CONSTRAINT FK_EAF1E07DA76ED395 FOREIGN KEY (user_id) REFERENCES user (id) NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('INSERT INTO user_spreadsheet (id, user_id, spreadsheet_id, title, month, year) SELECT id, user_id, spreadsheet_id, title, month, year FROM __temp__user_spreadsheet');
        $this->addSql('DROP TABLE __temp__user_spreadsheet');
        $this->addSql('CREATE INDEX IDX_EAF1E07DA76ED395 ON user_spreadsheet (user_id)');
    }
}
