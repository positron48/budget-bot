<?php

namespace App\Tests\Integration\Command;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserRepository;
use App\Repository\UserSpreadsheetRepository;
use App\Tests\Integration\AbstractBotIntegrationTestCase;

class RemoveCommandTest extends AbstractBotIntegrationTestCase
{
    private UserRepository $userRepository;
    private UserSpreadsheetRepository $spreadsheetRepository;

    protected function setUp(): void
    {
        parent::setUp();
        $this->userRepository = $this->getContainer()->get(UserRepository::class);
        $this->spreadsheetRepository = $this->getContainer()->get(UserSpreadsheetRepository::class);
    }

    public function testExecuteWithoutStartCommand(): void
    {
        $this->executeCommand('/remove', 123456);

        $this->assertMessageCount(1);
        $this->assertLastMessageContains('Пожалуйста, начните с команды /start');
    }

    public function testExecuteWithNoSpreadsheets(): void
    {
        // Create a user
        $user = new User();
        $user->setTelegramId(123456);
        $this->userRepository->save($user, true);

        $this->executeCommand('/remove', 123456);

        $this->assertMessageCount(1);
        $this->assertLastMessageContains('У вас пока нет добавленных таблиц');
    }

    public function testExecuteWithSpreadsheets(): void
    {
        // Create a user with spreadsheets
        $user = new User();
        $user->setTelegramId(123456);
        $this->userRepository->save($user, true);

        // Add test spreadsheets
        $spreadsheet1 = new UserSpreadsheet();
        $spreadsheet1->setUser($user);
        $spreadsheet1->setSpreadsheetId('test_id_1');
        $spreadsheet1->setMonth(1);
        $spreadsheet1->setYear(2024);
        $spreadsheet1->setTitle('Test Budget 1 2024');
        $this->spreadsheetRepository->save($spreadsheet1, true);

        $spreadsheet2 = new UserSpreadsheet();
        $spreadsheet2->setUser($user);
        $spreadsheet2->setSpreadsheetId('test_id_2');
        $spreadsheet2->setMonth(2);
        $spreadsheet2->setYear(2024);
        $spreadsheet2->setTitle('Test Budget 2 2024');
        $this->spreadsheetRepository->save($spreadsheet2, true);

        // Setup test spreadsheet in Google API mock
        $this->setupTestSpreadsheet('test_id_1');
        $this->setupTestSpreadsheet('test_id_2');
        $this->googleApiClient->setSpreadsheetTitle('test_id_1', 'Test Budget 1 2024');
        $this->googleApiClient->setSpreadsheetTitle('test_id_2', 'Test Budget 2 2024');
        $this->googleApiClient->setSpreadsheetAccessible('test_id_1', true);
        $this->googleApiClient->setSpreadsheetAccessible('test_id_2', true);

        // Test listing spreadsheets
        $this->executeCommand('/remove', 123456);

        $this->assertMessageCount(1);
        $this->assertLastMessageContains('Выберите таблицу для удаления');

        // Verify user state is set
        $this->entityManager->clear();
        $user = $this->userRepository->findByTelegramId(123456);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_REMOVE_SPREADSHEET', $user->getState());

        // Test removing specific spreadsheet
        $this->executeCommand('/remove Январь 2024', 123456);

        $this->assertMessageCount(2);
        $this->assertLastMessageContains('Таблица успешно удалена');

        // Verify spreadsheet is removed
        $this->entityManager->clear();
        $remainingSpreadsheets = $this->spreadsheetRepository->findBy(['user' => $user]);
        $this->assertCount(1, $remainingSpreadsheets);
        $this->assertEquals('test_id_2', $remainingSpreadsheets[0]->getSpreadsheetId());
    }

    public function testExecuteWithNonExistentSpreadsheet(): void
    {
        // Create a user with spreadsheets
        $user = new User();
        $user->setTelegramId(123456);
        $this->userRepository->save($user, true);

        // Add test spreadsheet
        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setUser($user);
        $spreadsheet->setSpreadsheetId('test_id_1');
        $spreadsheet->setMonth(1);
        $spreadsheet->setYear(2024);
        $spreadsheet->setTitle('Test Budget 1 2024');
        $this->spreadsheetRepository->save($spreadsheet, true);

        // Setup test spreadsheet in Google API mock
        $this->setupTestSpreadsheet('test_id_1');
        $this->googleApiClient->setSpreadsheetTitle('test_id_1', 'Test Budget 1 2024');
        $this->googleApiClient->setSpreadsheetAccessible('test_id_1', true);

        // Try to remove non-existent spreadsheet
        $this->executeCommand('/remove Февраль 2024', 123456);

        $this->assertMessageCount(1);
        $this->assertLastMessageContains('Таблица не найдена');

        // Verify original spreadsheet still exists
        $this->entityManager->clear();
        $remainingSpreadsheets = $this->spreadsheetRepository->findBy(['user' => $user]);
        $this->assertCount(1, $remainingSpreadsheets);
        $this->assertEquals('test_id_1', $remainingSpreadsheets[0]->getSpreadsheetId());
    }

    public function testExecuteWithInvalidSpreadsheetAccess(): void
    {
        // Create a user with spreadsheets
        $user = new User();
        $user->setTelegramId(123456);
        $this->userRepository->save($user, true);

        // Add test spreadsheet
        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setUser($user);
        $spreadsheet->setSpreadsheetId('test_id_1');
        $spreadsheet->setMonth(1);
        $spreadsheet->setYear(2024);
        $spreadsheet->setTitle('Test Budget 1 2024');
        $this->spreadsheetRepository->save($spreadsheet, true);

        // Setup test spreadsheet in Google API mock as inaccessible
        $this->setupTestSpreadsheet('test_id_1');
        $this->googleApiClient->setSpreadsheetTitle('test_id_1', 'Test Budget 1 2024');
        $this->googleApiClient->setSpreadsheetAccessible('test_id_1', false);

        // Try to remove spreadsheet
        $this->executeCommand('/remove Январь 2024', 123456);

        $this->assertMessageCount(1);
        $this->assertLastMessageContains('Не удалось получить доступ к таблице');

        // Verify spreadsheet still exists
        $this->entityManager->clear();
        $remainingSpreadsheets = $this->spreadsheetRepository->findBy(['user' => $user]);
        $this->assertCount(1, $remainingSpreadsheets);
        $this->assertEquals('test_id_1', $remainingSpreadsheets[0]->getSpreadsheetId());
    }
}
