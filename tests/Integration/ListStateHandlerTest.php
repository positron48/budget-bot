<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;
use App\Service\StateHandler\ListStateHandler;

class ListStateHandlerTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private const TEST_SPREADSHEET_ID = 'test_spreadsheet';

    private UserRepository $userRepository;
    private ListStateHandler $listStateHandler;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);
        $this->listStateHandler = $container->get(ListStateHandler::class);
    }

    public function testSupportsMethod(): void
    {
        $this->assertTrue($this->listStateHandler->supports('WAITING_LIST_ACTION'));
        $this->assertFalse($this->listStateHandler->supports('INVALID_STATE'));
    }

    public function testHandleWithInvalidAction(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'InvalidAction');
        $this->assertFalse($result);
    }

    public function testHandleWithMissingTempData(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([]); // Empty temp data
        $this->userRepository->save($user, true);

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Missing required temp data');

        $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
    }

    public function testHandleWithNoTransactions(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock empty values from Google Sheets
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!B5:E', []);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
        $this->assertTrue($result);

        // Verify message
        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Нет расходов за Январь 2025',
            $lastMessages[count($lastMessages) - 1]['text']
        );

        // Verify state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEmpty($updatedUser->getState());
        $this->assertEmpty($updatedUser->getTempData());
    }

    public function testHandleWithTransactions(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock transactions from Google Sheets
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!B5:E', [
            ['01.01.2025', '1500', 'Продукты', 'Питание'],
            ['02.01.2025', '2000', 'Такси', 'Транспорт'],
            ['15.01.2025', '3000', 'Ресторан', 'Питание'],
        ]);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
        $this->assertTrue($result);

        // Verify message
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1]['text'];

        $this->assertStringContainsString('Расходы за Январь 2025 (страница 1 из 1):', $lastMessage);
        $this->assertStringContainsString('01.01.2025 | 1 500.00 руб. | [Питание] Продукты', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('15.01.2025 | 3 000.00 руб. | [Питание] Ресторан', $lastMessage);
        $this->assertStringContainsString('Итого за страницу: 6500.00 руб.', $lastMessage);
        $this->assertStringContainsString('Общий итог: 6500.00 руб.', $lastMessage);

        // Verify state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEquals('WAITING_LIST_PAGE', $updatedUser->getState());
        $tempData = $updatedUser->getTempData();
        $this->assertEquals(1, $tempData['current_page']);
        $this->assertEquals('Расходы', $tempData['type']);
        $this->assertCount(3, $tempData['transactions']);
    }

    public function testHandleWithInvalidTransactionDate(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock transactions with invalid date
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!B5:E', [
            ['invalid_date', '1500', 'Продукты', 'Питание'],
            ['02.01.2025', '2000', 'Такси', 'Транспорт'],
        ]);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
        $this->assertTrue($result);

        // Verify message (should only include valid transaction)
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1]['text'];

        $this->assertStringContainsString('Расходы за Январь 2025 (страница 1 из 1):', $lastMessage);
        $this->assertStringNotContainsString('invalid_date', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('Итого за страницу: 2000.00 руб.', $lastMessage);
        $this->assertStringContainsString('Общий итог: 2000.00 руб.', $lastMessage);
    }

    public function testHandleWithIncompleteTransactionData(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock transactions with incomplete data
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!B5:E', [
            ['01.01.2025', '1500'], // Missing description and category
            ['02.01.2025', '2000', 'Такси', 'Транспорт'],
        ]);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
        $this->assertTrue($result);

        // Verify message (should only include complete transaction)
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1]['text'];

        $this->assertStringContainsString('Расходы за Январь 2025 (страница 1 из 1):', $lastMessage);
        $this->assertStringNotContainsString('01.01.2025 | 1 500', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('Итого за страницу: 2000.00 руб.', $lastMessage);
        $this->assertStringContainsString('Общий итог: 2000.00 руб.', $lastMessage);
    }

    public function testHandleWithIncome(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock income transactions from Google Sheets (using the correct range G5:J)
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!G5:J', [
            ['01.01.2025', '50000', 'Зарплата', 'Доход'],
            ['15.01.2025', '10000', 'Подработка', 'Доход'],
        ]);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Доходы');
        $this->assertTrue($result);

        // Verify message
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1]['text'];

        $this->assertStringContainsString('Доходы за Январь 2025 (страница 1 из 1):', $lastMessage);
        $this->assertStringContainsString('01.01.2025 | 50 000.00 руб. | [Доход] Зарплата', $lastMessage);
        $this->assertStringContainsString('15.01.2025 | 10 000.00 руб. | [Доход] Подработка', $lastMessage);
        $this->assertStringContainsString('Итого за страницу: 60000.00 руб.', $lastMessage);
        $this->assertStringContainsString('Общий итог: 60000.00 руб.', $lastMessage);

        // Verify state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEquals('WAITING_LIST_PAGE', $updatedUser->getState());
        $tempData = $updatedUser->getTempData();
        $this->assertEquals(1, $tempData['current_page']);
        $this->assertEquals('Доходы', $tempData['type']);
        $this->assertCount(2, $tempData['transactions']);
    }

    public function testHandleWithTransactionsFromDifferentMonths(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock transactions from different months
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!B5:E', [
            ['01.01.2025', '1500', 'Продукты', 'Питание'],
            ['02.02.2025', '2000', 'Такси', 'Транспорт'], // Different month
            ['15.01.2025', '3000', 'Ресторан', 'Питание'],
        ]);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
        $this->assertTrue($result);

        // Verify message (should only include transactions from January)
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1]['text'];

        $this->assertStringContainsString('Расходы за Январь 2025 (страница 1 из 1):', $lastMessage);
        $this->assertStringContainsString('01.01.2025 | 1 500.00 руб. | [Питание] Продукты', $lastMessage);
        $this->assertStringContainsString('15.01.2025 | 3 000.00 руб. | [Питание] Ресторан', $lastMessage);
        $this->assertStringNotContainsString('02.02.2025', $lastMessage);
        $this->assertStringContainsString('Итого за страницу: 4500.00 руб.', $lastMessage);
        $this->assertStringContainsString('Общий итог: 4500.00 руб.', $lastMessage);
    }

    public function testHandleWithInvalidAmount(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock transactions with invalid amount
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!B5:E', [
            ['01.01.2025', 'not_a_number', 'Продукты', 'Питание'],
            ['02.01.2025', '2000', 'Такси', 'Транспорт'],
        ]);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
        $this->assertTrue($result);

        // Verify message (should only include valid transaction)
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1]['text'];

        $this->assertStringContainsString('Расходы за Январь 2025 (страница 1 из 1):', $lastMessage);
        $this->assertStringNotContainsString('not_a_number', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('Итого за страницу: 2000.00 руб.', $lastMessage);
        $this->assertStringContainsString('Общий итог: 2000.00 руб.', $lastMessage);
    }

    public function testHandleWithPagination(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => 1,
            'list_year' => 2025,
            'spreadsheet_id' => self::TEST_SPREADSHEET_ID,
        ]);
        $this->userRepository->save($user, true);

        // Mock transactions (more than TRANSACTIONS_PER_PAGE)
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Транзакции!B5:E', [
            ['01.01.2025', '1500', 'Продукты', 'Питание'],
            ['02.01.2025', '2000', 'Такси', 'Транспорт'],
            ['03.01.2025', '3000', 'Ресторан', 'Питание'],
            ['04.01.2025', '1000', 'Метро', 'Транспорт'],
            ['05.01.2025', '2500', 'Кино', 'Развлечения'],
            ['06.01.2025', '1800', 'Продукты', 'Питание'],
            ['07.01.2025', '4000', 'Одежда', 'Покупки'],
        ]);

        // Initial request
        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, 'Расходы');
        $this->assertTrue($result);

        // Verify first page
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1];
        $text = $lastMessage['text'];
        $keyboard = json_decode($lastMessage['reply_markup'], true)['keyboard'];

        $this->assertStringContainsString('Расходы за Январь 2025 (страница 1 из 2):', $text);
        $this->assertStringContainsString('07.01.2025 | 4 000.00 руб. | [Покупки] Одежда', $text);
        $this->assertStringContainsString('06.01.2025 | 1 800.00 руб. | [Питание] Продукты', $text);
        $this->assertStringContainsString('05.01.2025 | 2 500.00 руб. | [Развлечения] Кино', $text);
        $this->assertStringContainsString('04.01.2025 | 1 000.00 руб. | [Транспорт] Метро', $text);
        $this->assertStringContainsString('03.01.2025 | 3 000.00 руб. | [Питание] Ресторан', $text);
        $this->assertStringContainsString('Итого за страницу: 12300.00 руб.', $text);
        $this->assertStringContainsString('Общий итог: 15800.00 руб.', $text);

        // Verify keyboard buttons (navigation buttons on first page)
        $this->assertCount(2, $keyboard);
        $this->assertCount(1, $keyboard[0]); // ➡️ Вперед
        $this->assertEquals('➡️ Вперед', $keyboard[0][0]['text']);
        $this->assertCount(1, $keyboard[1]); // ❌ Закрыть
        $this->assertEquals('❌ Закрыть', $keyboard[1][0]['text']);

        // Navigate to next page
        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, '➡️ Вперед');
        $this->assertTrue($result);

        // Verify second page
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1];
        $text = $lastMessage['text'];
        $keyboard = json_decode($lastMessage['reply_markup'], true)['keyboard'];

        $this->assertStringContainsString('Расходы за Январь 2025 (страница 2 из 2):', $text);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $text);
        $this->assertStringContainsString('01.01.2025 | 1 500.00 руб. | [Питание] Продукты', $text);
        $this->assertStringContainsString('Итого за страницу: 3500.00 руб.', $text);
        $this->assertStringContainsString('Общий итог: 15800.00 руб.', $text);

        // Verify keyboard buttons (back button and close button on second page)
        $this->assertCount(2, $keyboard);
        $this->assertCount(1, $keyboard[0]); // ⬅️ Назад
        $this->assertEquals('⬅️ Назад', $keyboard[0][0]['text']);
        $this->assertCount(1, $keyboard[1]); // ❌ Закрыть
        $this->assertEquals('❌ Закрыть', $keyboard[1][0]['text']);

        // Close the list
        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, '❌ Закрыть');
        $this->assertTrue($result);

        // Verify state is cleared
        $finalUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($finalUser);
        $this->assertEmpty($finalUser->getState());
        $this->assertEmpty($finalUser->getTempData());

        // Verify close message
        $lastMessages = $this->telegramApi->getMessages();
        $lastMessage = $lastMessages[count($lastMessages) - 1];
        $this->assertEquals('Просмотр транзакций завершен', $lastMessage['text']);
        $this->assertTrue(json_decode($lastMessage['reply_markup'], true)['remove_keyboard']);
    }

    public function testHandleWithInvalidPageNavigation(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            $this->executeCommand('/start', self::TEST_CHAT_ID);
            $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        }
        $this->assertNotNull($user);

        // Set up state without required temp data
        $user->setState('WAITING_LIST_PAGE');
        $user->setTempData([]);
        $this->userRepository->save($user, true);

        $result = $this->listStateHandler->handle(self::TEST_CHAT_ID, $user, '➡️ Вперед');
        $this->assertFalse($result);

        // Verify state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEmpty($updatedUser->getState());
        $this->assertEmpty($updatedUser->getTempData());
    }
}
