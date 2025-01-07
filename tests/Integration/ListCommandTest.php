<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;
use App\Repository\UserSpreadsheetRepository;

class ListCommandTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private const TEST_SPREADSHEET_ID = 'test_spreadsheet';

    private UserRepository $userRepository;
    private UserSpreadsheetRepository $spreadsheetRepository;

    protected function setUp(): void
    {
        parent::setUp();
        $this->userRepository = self::getContainer()->get(UserRepository::class);
        $this->spreadsheetRepository = self::getContainer()->get(UserSpreadsheetRepository::class);
    }

    private function setupInitialState(): void
    {
        // Setup test spreadsheet first
        $this->setupTestSpreadsheet(self::TEST_SPREADSHEET_ID);
        $this->googleApiClient->setSpreadsheetAccessible(self::TEST_SPREADSHEET_ID, true);
        $this->googleApiClient->setSpreadsheetTitle(self::TEST_SPREADSHEET_ID, 'Test Budget');

        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add spreadsheet
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.self::TEST_SPREADSHEET_ID.'/edit', self::TEST_CHAT_ID);
        $this->executeCommand('Январь 2025', self::TEST_CHAT_ID);
    }

    public function testGetName(): void
    {
        $command = self::getContainer()->get('App\Service\Command\ListCommand');
        $this->assertEquals('/list', $command->getName());
    }

    public function testSupportsMethod(): void
    {
        $command = self::getContainer()->get('App\Service\Command\ListCommand');
        $this->assertTrue($command->supports('/list'));
        $this->assertTrue($command->supports('/list Январь'));
        $this->assertTrue($command->supports('/list Январь 2025'));
        $this->assertTrue($command->supports('  /list  ')); // with spaces
        $this->assertFalse($command->supports('/start'));
        $this->assertFalse($command->supports('/list_tables'));
    }

    public function testListWithoutUser(): void
    {
        // Try to list without starting the bot first
        $this->executeCommand('/list', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Пожалуйста, начните с команды /start',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testListWithoutSpreadsheet(): void
    {
        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Try to list without adding a spreadsheet
        $this->executeCommand('/list', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $now = new \DateTime();
        $this->assertStringContainsString(
            sprintf('У вас нет таблицы за %s %d', $this->getMonthName((int) $now->format('m')), (int) $now->format('Y')),
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testListCurrentMonth(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Execute list command
        $this->executeCommand('/list', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Выберите тип транзакций за',
            $lastMessages[count($lastMessages) - 1]['text']
        );

        // Verify keyboard options
        $lastMessage = $lastMessages[count($lastMessages) - 1];
        $this->assertArrayHasKey('reply_markup', $lastMessage);
        $replyMarkup = json_decode($lastMessage['reply_markup'], true);
        $this->assertNotNull($replyMarkup);
        $this->assertCount(2, $replyMarkup['keyboard']);
        $this->assertEquals('Расходы', $replyMarkup['keyboard'][0][0]['text']);
        $this->assertEquals('Доходы', $replyMarkup['keyboard'][1][0]['text']);

        // Verify state is set correctly
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_LIST_ACTION', $user->getState());
        $tempData = $user->getTempData();
        $this->assertIsArray($tempData);
        $this->assertArrayHasKey('list_month', $tempData);
        $this->assertArrayHasKey('list_year', $tempData);
        $this->assertArrayHasKey('spreadsheet_id', $tempData);
    }

    public function testListSpecificMonth(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Execute list command with specific month
        $this->executeCommand('/list Февраль 2025', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'У вас нет таблицы за Февраль 2025',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testListWithInvalidMonth(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Execute list command with invalid month
        $this->executeCommand('/list НеверныйМесяц 2025', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Неверный формат месяца. Пожалуйста, укажите месяц числом (1-12) или словом (Январь-Декабрь).',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testListWithInvalidYear(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Execute list command with invalid year
        $this->executeCommand('/list Январь НеверныйГод', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Неверный формат года. Пожалуйста, укажите год в числовом формате.',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testListWithYearBelow2020(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Execute list command with year before 2020
        $this->executeCommand('/list Январь 2019', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Год не может быть меньше 2020.',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testListWithNumericMonth(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Execute list command with numeric month
        $this->executeCommand('/list 1 2025', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Выберите тип транзакций за Январь 2025',
            $lastMessages[count($lastMessages) - 1]['text']
        );

        // Verify state is set correctly
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_LIST_ACTION', $user->getState());
        $tempData = $user->getTempData();
        $this->assertIsArray($tempData);
        $this->assertEquals(1, $tempData['list_month']);
        $this->assertEquals(2025, $tempData['list_year']);
    }

    public function testListWithInvalidNumericMonth(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Execute list command with invalid numeric month
        $this->executeCommand('/list 13 2025', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Неверный формат месяца. Пожалуйста, укажите месяц числом (1-12) или словом (Январь-Декабрь).',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testListWithNullSpreadsheetId(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Create a spreadsheet with null ID
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        if (!$user) {
            throw new \RuntimeException('User not found');
        }
        $spreadsheet = $this->spreadsheetRepository->findByMonthAndYear($user, 1, 2025);
        if (!$spreadsheet) {
            throw new \RuntimeException('Spreadsheet not found');
        }
        $spreadsheet->setSpreadsheetId('');
        $this->spreadsheetRepository->save($spreadsheet, true);

        // Execute list command
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Spreadsheet ID is null');

        $this->executeCommand('/list Январь 2025', self::TEST_CHAT_ID);
    }

    private function getMonthName(int $month): string
    {
        $months = [
            1 => 'Январь',
            2 => 'Февраль',
            3 => 'Март',
            4 => 'Апрель',
            5 => 'Май',
            6 => 'Июнь',
            7 => 'Июль',
            8 => 'Август',
            9 => 'Сентябрь',
            10 => 'Октябрь',
            11 => 'Ноябрь',
            12 => 'Декабрь',
        ];

        return $months[$month] ?? '';
    }
}
