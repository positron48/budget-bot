<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;
use App\Repository\UserSpreadsheetRepository;
use App\Service\Google\GoogleApiClientInterface;
use App\Tests\Integration\DataFixtures\TestFixtures;
use App\Tests\Mock\TestGoogleApiClient;
use Doctrine\Common\DataFixtures\Executor\ORMExecutor;
use Doctrine\Common\DataFixtures\Loader;
use Doctrine\Common\DataFixtures\Purger\ORMPurger;

class SpreadsheetStateFlowTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private const TEST_SPREADSHEET_ID = 'test_new_spreadsheet';
    private UserRepository $userRepository;
    private UserSpreadsheetRepository $spreadsheetRepository;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);
        $this->spreadsheetRepository = $container->get(UserSpreadsheetRepository::class);

        // Load test fixtures using DoctrineFixturesBundle
        $loader = new Loader();
        $loader->addFixture(new TestFixtures());

        $executor = new ORMExecutor($this->entityManager, new ORMPurger());
        $executor->execute($loader->getFixtures());

        // Setup test spreadsheet
        $this->setupTestSpreadsheet(self::TEST_SPREADSHEET_ID);
        $this->setupTestCategories(self::TEST_SPREADSHEET_ID);

        // Setup test spreadsheet in Google API client
        /** @var TestGoogleApiClient $client */
        $client = self::getContainer()->get(GoogleApiClientInterface::class);
        $client->setSpreadsheetAccessible(self::TEST_SPREADSHEET_ID, true);
        $client->setSpreadsheetTitle(self::TEST_SPREADSHEET_ID, 'Test Budget');
    }

    protected function setupTestSpreadsheet(string $spreadsheetId, ?string $title = null): void
    {
        /** @var TestGoogleApiClient $client */
        $client = self::getContainer()->get(GoogleApiClientInterface::class);
        $client->setSpreadsheetAccessible($spreadsheetId, true);
        $client->setSpreadsheetTitle($spreadsheetId, $title ?? 'Test Budget');
    }

    public function testSpreadsheetActionFlow(): void
    {
        // Start with user creation
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Check add command
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Отправьте ссылку на таблицу');

        // Verify user state
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_SPREADSHEET_ID', $user->getState());

        // Test invalid action
        $this->executeCommand('Неизвестное действие', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Неверный ID таблицы');
    }

    public function testAddSpreadsheetFlow(): void
    {
        // Start with user creation
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add command
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Отправьте ссылку на таблицу');

        // Verify user state
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_SPREADSHEET_ID', $user->getState());

        // Test invalid spreadsheet ID
        $this->executeCommand('invalid_id', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Неверный ID таблицы');

        // Send valid spreadsheet ID
        $spreadsheetId = self::TEST_SPREADSHEET_ID;
        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.$spreadsheetId, self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите месяц и год');
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_SPREADSHEET_MONTH', $user->getState());

        // Test invalid month format
        $this->executeCommand('Неверный формат', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Неверный формат');

        // Test invalid month name
        $this->executeCommand('НеверныйМесяц 2024', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Неверный формат');

        // Test invalid year
        $this->executeCommand('Январь 1999', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Неверный формат');

        // Send valid month and year
        $this->executeCommand('Март 2025', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Таблица за Март 2025 успешно добавлена');
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEmpty($user->getState());

        // Verify spreadsheet was saved
        $spreadsheet = $this->spreadsheetRepository->findByMonthAndYear($user, 3, 2025);
        $this->assertNotNull($spreadsheet);
        $this->assertEquals(3, $spreadsheet->getMonth());
        $this->assertEquals(2025, $spreadsheet->getYear());
    }

    public function testDeleteSpreadsheetFlow(): void
    {
        // First add a spreadsheet
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add command
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Отправьте ссылку на таблицу');

        // Send valid spreadsheet ID
        $spreadsheetId = self::TEST_SPREADSHEET_ID;
        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.$spreadsheetId, self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите месяц и год');

        // Send valid month and year
        $this->executeCommand('Март 2025', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Таблица за Март 2025 успешно добавлена');

        // Start delete flow
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Отправьте ссылку на таблицу');

        // Set state for delete action
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $user->setState('WAITING_SPREADSHEET_ACTION');
        $this->userRepository->save($user, true);

        // Send delete action
        $this->executeCommand('Удалить таблицу', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите таблицу для удаления');

        // Verify user state
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_SPREADSHEET_TO_DELETE', $user->getState());

        // Test invalid spreadsheet selection
        $this->executeCommand('Неверная таблица', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Таблица не найдена');

        // Select valid spreadsheet
        $this->executeCommand('Март 2025', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Таблица за Март 2025 успешно удалена');
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEmpty($user->getState());

        // Verify spreadsheet was deleted
        $spreadsheet = $this->spreadsheetRepository->findByMonthAndYear($user, 3, 2025);
        $this->assertNull($spreadsheet);

        // Clear all spreadsheets
        $spreadsheets = $this->spreadsheetRepository->findBy(['user' => $user]);
        foreach ($spreadsheets as $spreadsheet) {
            $this->spreadsheetRepository->remove($spreadsheet, true);
        }

        // Try to delete when no spreadsheets exist
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Отправьте ссылку на таблицу');

        // Set state for delete action
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $user->setState('WAITING_SPREADSHEET_ACTION');
        $this->userRepository->save($user, true);

        // Send delete action
        $this->executeCommand('Удалить таблицу', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('У вас нет добавленных таблиц');
    }
}
