<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20250106182639 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE TABLE spreadsheet (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, user_id INTEGER NOT NULL, spreadsheet_id VARCHAR(255) NOT NULL, month INTEGER NOT NULL, year INTEGER NOT NULL, CONSTRAINT FK_43EA29E8A76ED395 FOREIGN KEY (user_id) REFERENCES user (id) NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('CREATE INDEX IDX_43EA29E8A76ED395 ON spreadsheet (user_id)');
        $this->addSql('ALTER TABLE user_category ADD COLUMN is_income BOOLEAN NOT NULL DEFAULT FALSE');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('DROP TABLE spreadsheet');
        $this->addSql('CREATE TEMPORARY TABLE __temp__user_category AS SELECT id, user_id, name, type FROM user_category');
        $this->addSql('DROP TABLE user_category');
        $this->addSql('CREATE TABLE user_category (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, user_id INTEGER NOT NULL, name VARCHAR(255) NOT NULL, type VARCHAR(10) NOT NULL, CONSTRAINT FK_E6C1FDC1A76ED395 FOREIGN KEY (user_id) REFERENCES user (id) NOT DEFERRABLE INITIALLY IMMEDIATE)');
        $this->addSql('INSERT INTO user_category (id, user_id, name, type) SELECT id, user_id, name, type FROM __temp__user_category');
        $this->addSql('DROP TABLE __temp__user_category');
        $this->addSql('CREATE INDEX IDX_E6C1FDC1A76ED395 ON user_category (user_id)');
        $this->addSql('CREATE UNIQUE INDEX UNIQ_E9B0D9895E237E06979B1AD6A76ED395 ON user_category (name, type, user_id)');
    }
}
