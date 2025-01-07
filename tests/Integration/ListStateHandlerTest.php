<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;
use App\Service\StateHandler\ListStateHandler;
use Psr\Log\LoggerInterface;

class ListStateHandlerTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private const TEST_SPREADSHEET_ID = 'test_spreadsheet';

    private UserRepository $userRepository;
    private ListStateHandler $listStateHandler;
    private LoggerInterface $logger;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);
        $this->logger = $container->get(LoggerInterface::class);

        $this->listStateHandler = new ListStateHandler(
            $this->userRepository,
            $this->telegramApi,
            $this->googleApiClient,
            $this->logger
        );
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

        $this->assertStringContainsString('Расходы за Январь 2025:', $lastMessage);
        $this->assertStringContainsString('01.01.2025 | 1 500.00 руб. | [Питание] Продукты', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('15.01.2025 | 3 000.00 руб. | [Питание] Ресторан', $lastMessage);
        $this->assertStringContainsString('Итого: 6500.00 руб.', $lastMessage);

        // Verify state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEmpty($updatedUser->getState());
        $this->assertEmpty($updatedUser->getTempData());
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

        $this->assertStringContainsString('Расходы за Январь 2025:', $lastMessage);
        $this->assertStringNotContainsString('invalid_date', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('Итого: 2000.00 руб.', $lastMessage);
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

        $this->assertStringContainsString('Расходы за Январь 2025:', $lastMessage);
        $this->assertStringNotContainsString('01.01.2025 | 1 500', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('Итого: 2000.00 руб.', $lastMessage);
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

        $this->assertStringContainsString('Доходы за Январь 2025:', $lastMessage);
        $this->assertStringContainsString('01.01.2025 | 50 000.00 руб. | [Доход] Зарплата', $lastMessage);
        $this->assertStringContainsString('15.01.2025 | 10 000.00 руб. | [Доход] Подработка', $lastMessage);
        $this->assertStringContainsString('Итого: 60000.00 руб.', $lastMessage);

        // Verify state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEmpty($updatedUser->getState());
        $this->assertEmpty($updatedUser->getTempData());
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

        $this->assertStringContainsString('Расходы за Январь 2025:', $lastMessage);
        $this->assertStringContainsString('01.01.2025 | 1 500.00 руб. | [Питание] Продукты', $lastMessage);
        $this->assertStringContainsString('15.01.2025 | 3 000.00 руб. | [Питание] Ресторан', $lastMessage);
        $this->assertStringNotContainsString('02.02.2025', $lastMessage);
        $this->assertStringContainsString('Итого: 4500.00 руб.', $lastMessage);
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

        $this->assertStringContainsString('Расходы за Январь 2025:', $lastMessage);
        $this->assertStringNotContainsString('not_a_number', $lastMessage);
        $this->assertStringContainsString('02.01.2025 | 2 000.00 руб. | [Транспорт] Такси', $lastMessage);
        $this->assertStringContainsString('Итого: 2000.00 руб.', $lastMessage);
    }
}
