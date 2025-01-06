<?php

namespace App\Tests\Integration;

use Doctrine\ORM\EntityManager;
use Doctrine\ORM\Mapping\ClassMetadata;
use Doctrine\ORM\Tools\SchemaTool;
use Google\Service\Drive;
use Google\Service\Sheets;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Telegram;
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

        // Reset database
        $connection = $entityManager->getConnection();
        $platform = $connection->getDatabasePlatform();

        $tables = $connection->createSchemaManager()->listTableNames();
        foreach ($tables as $table) {
            $connection->executeStatement($platform->getTruncateTableSQL($table, true));
        }

        // Create database schema
        $schemaTool = new SchemaTool($entityManager);
        /** @var array<ClassMetadata<object>> $metadata */
        $metadata = $entityManager->getMetadataFactory()->getAllMetadata();
        $schemaTool->createSchema($metadata);

        // Mock Telegram API
        $this->mockTelegramApi();

        // Mock Google services
        $this->mockGoogleServices();
    }

    protected function tearDown(): void
    {
        parent::tearDown();

        if (isset($this->entityManager)) {
            $this->entityManager->close();
            unset($this->entityManager);
        }
    }

    private function mockTelegramApi(): void
    {
        if (!function_exists('runkit7_method_redefine')) {
            $this->markTestSkipped('runkit extension is required for this test');
        }

        if (!defined('RUNKIT7_ACC_PUBLIC')) {
            define('RUNKIT7_ACC_PUBLIC', 1);
        }
        if (!defined('RUNKIT7_ACC_STATIC')) {
            define('RUNKIT7_ACC_STATIC', 4);
        }
        if (!defined('RUNKIT7_ACC_PRIVATE')) {
            define('RUNKIT7_ACC_PRIVATE', 2);
        }

        runkit7_method_redefine(
            Telegram::class,
            '__construct',
            '$api_key, $bot_username = ""',
            'return;',
            RUNKIT7_ACC_PUBLIC
        );

        runkit7_method_redefine(
            Request::class,
            'initialize',
            '$telegram',
            'return;',
            RUNKIT7_ACC_STATIC | RUNKIT7_ACC_PUBLIC
        );

        runkit7_method_redefine(
            Request::class,
            'sendMessage',
            '$data',
            'error_log("Telegram API Request: " . json_encode($data, JSON_UNESCAPED_UNICODE)); 
            if (is_array($data) && isset($data["text"])) { 
                $text = $data["text"];
                if (isset($data["reply_markup"])) {
                    $keyboard = $data["reply_markup"];
                    if (isset($keyboard["keyboard"])) {
                        $text .= "\nКлавиатура:\n";
                        foreach ($keyboard["keyboard"] as $row) {
                            foreach ($row as $button) {
                                $text .= "- " . $button["text"] . "\n";
                            }
                        }
                    }
                }
                \App\Tests\Mock\ResponseCollector::getInstance()->addResponse($text);
            } 
            $response = new \App\Tests\Mock\ServerResponseMock(["ok" => true]); 
            error_log("Telegram API Response: " . json_encode($response, JSON_UNESCAPED_UNICODE)); 
            return $response;',
            RUNKIT7_ACC_STATIC | RUNKIT7_ACC_PUBLIC
        );
    }

    private function mockGoogleServices(): void
    {
        if (!function_exists('runkit7_method_redefine')) {
            $this->markTestSkipped('runkit extension is required for this test');
        }

        if (!defined('RUNKIT7_ACC_PUBLIC')) {
            define('RUNKIT7_ACC_PUBLIC', 1);
        }

        runkit7_method_redefine(
            \Google\Client::class,
            'setAuthConfig',
            '$config',
            'return $this;',
            RUNKIT7_ACC_PUBLIC
        );

        runkit7_method_redefine(
            \Google\Client::class,
            'setScopes',
            '$scopes',
            'return $this;',
            RUNKIT7_ACC_PUBLIC
        );

        runkit7_method_redefine(
            Sheets::class,
            '__construct',
            '$client',
            'return;',
            RUNKIT7_ACC_PUBLIC
        );

        runkit7_method_redefine(
            Drive::class,
            '__construct',
            '$client',
            'return;',
            RUNKIT7_ACC_PUBLIC
        );
    }
}
