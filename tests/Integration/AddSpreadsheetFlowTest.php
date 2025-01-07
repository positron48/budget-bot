<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;
use App\Repository\UserSpreadsheetRepository;

class AddSpreadsheetFlowTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private const TEST_SPREADSHEET_ID = '1234567890';

    private UserRepository $userRepository;
    private UserSpreadsheetRepository $spreadsheetRepository;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);
        $this->spreadsheetRepository = $container->get(UserSpreadsheetRepository::class);

        // Set up test data
        $this->setupTestSpreadsheet(self::TEST_SPREADSHEET_ID);
    }

    public function testUserCreationOnStart(): void
    {
        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Verify welcome message
        $this->assertMessageCount(1);
        $this->assertLastMessageContains('Привет!');

        // Verify user creation
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
    }

    public function testAddSpreadsheetFlow(): void
    {
        // Start with user creation
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add command
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->assertMessageCount(2);
        $this->assertLastMessageContains('Отправьте ссылку на таблицу');

        // Verify user state
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_SPREADSHEET_ID', $user->getState());

        // Send spreadsheet link
        $spreadsheetLink = 'https://docs.google.com/spreadsheets/d/'.self::TEST_SPREADSHEET_ID.'/edit';
        $this->executeCommand($spreadsheetLink, self::TEST_CHAT_ID);
        $this->assertMessageCount(3);
        $this->assertLastMessageContains('Выберите месяц');

        // Select month
        $this->executeCommand('Январь 2024', self::TEST_CHAT_ID);
        $this->assertMessageCount(4);
        $this->assertLastMessageContains('Таблица за Январь 2024 успешно добавлена');

        // Verify final user state
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertTrue(null === $user->getState() || '' === $user->getState(), 'State should be null or empty string');

        // Verify spreadsheet was saved
        $spreadsheet = $this->spreadsheetRepository->findOneBy(['user' => $user]);
        $this->assertNotNull($spreadsheet);
        $this->assertEquals(self::TEST_SPREADSHEET_ID, $spreadsheet->getSpreadsheetId());
        $this->assertEquals(1, $spreadsheet->getMonth());
        $this->assertEquals(2024, $spreadsheet->getYear());
    }

    public function testListSpreadsheets(): void
    {
        // Setup spreadsheet
        $this->testAddSpreadsheetFlow();

        // Check list command
        $this->executeCommand('/list', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Январь 2024');
    }

    public function testRemoveSpreadsheet(): void
    {
        // Setup spreadsheet
        $this->testAddSpreadsheetFlow();

        // Remove command
        $this->executeCommand('/remove', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите таблицу для удаления');

        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user, 'User should exist at this point');
        $this->assertSame('WAITING_REMOVE_SPREADSHEET', $user->getState());

        // Select spreadsheet to delete
        $this->executeCommand('Январь 2024', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Таблица за Январь 2024 успешно удалена');

        // Verify spreadsheet was deleted
        $spreadsheet = $this->spreadsheetRepository->findOneBy(['user' => $user]);
        $this->assertNull($spreadsheet);

        // Verify empty list
        $this->executeCommand('/list', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('У вас пока нет добавленных таблиц');
    }
}
