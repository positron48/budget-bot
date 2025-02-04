<?php

namespace App\Tests\Integration;

use Doctrine\ORM\EntityManager;
use Doctrine\ORM\Mapping\ClassMetadata;
use Doctrine\ORM\Tools\SchemaTool;
use Symfony\Bundle\FrameworkBundle\Test\KernelTestCase;

abstract class IntegrationTestCase extends KernelTestCase
{
    protected EntityManager $entityManager;

    protected function setUp(): void
    {
        self::bootKernel();

        $container = static::getContainer();
        /** @var EntityManager $entityManager */
        $entityManager = $container->get('doctrine')->getManager();
        $this->entityManager = $entityManager;

        // Drop and recreate database schema
        $schemaTool = new SchemaTool($entityManager);
        /** @var array<ClassMetadata<object>> $metadata */
        $metadata = $entityManager->getMetadataFactory()->getAllMetadata();

        $connection = $entityManager->getConnection();
        $platform = $connection->getDatabasePlatform()->getName();

        // Handle foreign key constraints based on database platform
        if ('mysql' === $platform) {
            $connection->executeStatement('SET FOREIGN_KEY_CHECKS = 0');
        } elseif ('sqlite' === $platform) {
            $connection->executeStatement('PRAGMA foreign_keys = OFF');
        }

        try {
            $schemaTool->dropSchema($metadata);
        } catch (\Exception $e) {
            // Ignore errors during schema drop
        }

        // Clear entity manager to ensure clean state
        $entityManager->clear();

        try {
            $schemaTool->createSchema($metadata);
        } catch (\Exception $e) {
            throw new \RuntimeException('Failed to create database schema: '.$e->getMessage(), 0, $e);
        }

        // Re-enable foreign key constraints
        if ('mysql' === $platform) {
            $connection->executeStatement('SET FOREIGN_KEY_CHECKS = 1');
        } elseif ('sqlite' === $platform) {
            $connection->executeStatement('PRAGMA foreign_keys = ON');
        }
    }

    protected function tearDown(): void
    {
        parent::tearDown();

        if (isset($this->entityManager)) {
            $this->entityManager->close();
            unset($this->entityManager);
        }
    }
}
