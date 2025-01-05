<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20250105175406 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Remove duplicate categories and add unique constraints';
    }

    public function up(Schema $schema): void
    {
        // Remove duplicate default categories
        $this->addSql('
            WITH duplicates AS (
                SELECT name, type, MIN(id) as min_id
                FROM category
                GROUP BY name, type
                HAVING COUNT(*) > 1
            )
            DELETE FROM category
            WHERE id IN (
                SELECT c.id
                FROM category c
                JOIN duplicates d ON c.name = d.name AND c.type = d.type
                WHERE c.id > d.min_id
            )
        ');

        // Remove duplicate user categories
        $this->addSql('
            WITH duplicates AS (
                SELECT name, type, user_id, MIN(id) as min_id
                FROM user_category
                GROUP BY name, type, user_id
                HAVING COUNT(*) > 1
            )
            DELETE FROM user_category
            WHERE id IN (
                SELECT uc.id
                FROM user_category uc
                JOIN duplicates d ON uc.name = d.name AND uc.type = d.type AND uc.user_id = d.user_id
                WHERE uc.id > d.min_id
            )
        ');

        // Add unique constraints
        $this->addSql('CREATE UNIQUE INDEX UNIQ_64C19C15E237E06979B1AD6 ON category (name, type)');
        $this->addSql('CREATE UNIQUE INDEX UNIQ_E9B0D9895E237E06979B1AD6A76ED395 ON user_category (name, type, user_id)');
    }

    public function down(Schema $schema): void
    {
        // Remove unique constraints
        $this->addSql('DROP INDEX UNIQ_64C19C15E237E06979B1AD6 ON category');
        $this->addSql('DROP INDEX UNIQ_E9B0D9895E237E06979B1AD6A76ED395 ON user_category');
    }
}
