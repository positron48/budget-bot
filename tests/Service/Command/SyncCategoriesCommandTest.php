<?php

namespace App\Tests\Service\Command;

use App\Entity\Spreadsheet;
use App\Entity\User;
use App\Service\CategoryService;
use App\Service\Command\SyncCategoriesCommand;
use App\Service\GoogleSheetsService;
use App\Service\TelegramApiServiceInterface;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class SyncCategoriesCommandTest extends TestCase
{
    private SyncCategoriesCommand $command;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var GoogleSheetsService&MockObject */
    private GoogleSheetsService $sheetsService;
    /** @var CategoryService&MockObject */
    private CategoryService $categoryService;
    /** @var TelegramApiServiceInterface&MockObject */
    private TelegramApiServiceInterface $telegramApi;

    protected function setUp(): void
    {
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->sheetsService = $this->createMock(GoogleSheetsService::class);
        $this->categoryService = $this->createMock(CategoryService::class);
        $this->telegramApi = $this->createMock(TelegramApiServiceInterface::class);

        $this->command = new SyncCategoriesCommand(
            $this->logger,
            $this->sheetsService,
            $this->categoryService,
            $this->telegramApi
        );
    }

    public function testGetName(): void
    {
        $this->assertEquals('/sync_categories', $this->command->getName());
    }

    public function testSupports(): void
    {
        $this->assertTrue($this->command->supports('/sync_categories'));
        $this->assertFalse($this->command->supports('/start'));
    }

    public function testExecuteWithoutUser(): void
    {
        $chatId = 123456;

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, начните с команды /start',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, null, '/sync_categories');
    }

    public function testExecuteWithoutSpreadsheets(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->sheetsService->expects($this->once())
            ->method('findLatestSpreadsheet')
            ->with($user)
            ->willReturn(null);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/sync_categories');
    }

    public function testExecuteWithSpreadsheet(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $spreadsheet = new Spreadsheet();
        $spreadsheet->setSpreadsheetId('test-id');

        $this->sheetsService->expects($this->once())
            ->method('findLatestSpreadsheet')
            ->with($user)
            ->willReturn($spreadsheet);

        $this->categoryService->expects($this->once())
            ->method('getCategories')
            ->with(false, $user)
            ->willReturn(['Питание', 'Транспорт']);

        $this->categoryService->expects($this->once())
            ->method('getCategories')
            ->with(true, $user)
            ->willReturn(['Зарплата', 'Подработка']);

        $this->categoryService->expects($this->once())
            ->method('clearUserCategories')
            ->with($user);

        $this->sheetsService->expects($this->once())
            ->method('syncCategories')
            ->with($user, 'test-id')
            ->willReturn([
                'added_to_db' => [
                    'expense' => ['Развлечения'],
                    'income' => ['Инвестиции'],
                ],
                'added_to_sheet' => [
                    'expense' => ['Питание'],
                    'income' => ['Зарплата'],
                ],
            ]);

        $this->telegramApi->expects($this->exactly(2))
            ->method('sendMessage')
            ->willReturnCallback(function (array $data) use ($chatId) {
                static $callNumber = 0;
                ++$callNumber;

                if (1 === $callNumber) {
                    $this->assertEquals([
                        'chat_id' => $chatId,
                        'text' => "Пользовательские категории очищены:\n- Расходы: 2\n- Доходы: 2",
                        'parse_mode' => 'HTML',
                    ], $data);
                } elseif (2 === $callNumber) {
                    $this->assertEquals([
                        'chat_id' => $chatId,
                        'text' => "Синхронизация категорий завершена:\n\n".
                            "Добавлены в базу данных:\n".
                            "- Расходы: Развлечения\n".
                            "- Доходы: Инвестиции\n\n".
                            "Добавлены в таблицу:\n".
                            "- Расходы: Питание\n".
                            "- Доходы: Зарплата\n",
                        'parse_mode' => 'HTML',
                    ], $data);
                }
            });

        $this->command->execute($chatId, $user, '/sync_categories');
    }

    public function testExecuteWithSpreadsheetNoChanges(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $spreadsheet = new Spreadsheet();
        $spreadsheet->setSpreadsheetId('test-id');

        $this->sheetsService->expects($this->once())
            ->method('findLatestSpreadsheet')
            ->with($user)
            ->willReturn($spreadsheet);

        $this->categoryService->expects($this->once())
            ->method('getCategories')
            ->with(false, $user)
            ->willReturn(['Питание', 'Транспорт']);

        $this->categoryService->expects($this->once())
            ->method('getCategories')
            ->with(true, $user)
            ->willReturn(['Зарплата', 'Подработка']);

        $this->categoryService->expects($this->once())
            ->method('clearUserCategories')
            ->with($user);

        $this->sheetsService->expects($this->once())
            ->method('syncCategories')
            ->with($user, 'test-id')
            ->willReturn([
                'added_to_db' => [
                    'expense' => [],
                    'income' => [],
                ],
                'added_to_sheet' => [
                    'expense' => [],
                    'income' => [],
                ],
            ]);

        $this->telegramApi->expects($this->exactly(2))
            ->method('sendMessage')
            ->willReturnCallback(function (array $data) use ($chatId) {
                static $callNumber = 0;
                ++$callNumber;

                if (1 === $callNumber) {
                    $this->assertEquals([
                        'chat_id' => $chatId,
                        'text' => "Пользовательские категории очищены:\n- Расходы: 2\n- Доходы: 2",
                        'parse_mode' => 'HTML',
                    ], $data);
                } elseif (2 === $callNumber) {
                    $this->assertEquals([
                        'chat_id' => $chatId,
                        'text' => 'Все категории уже синхронизированы',
                        'parse_mode' => 'HTML',
                    ], $data);
                }
            });

        $this->command->execute($chatId, $user, '/sync_categories');
    }
}
