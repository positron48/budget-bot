<?php

namespace App\Tests\Service\Google;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use App\Service\Google\GoogleApiClientInterface;
use App\Service\Google\SpreadsheetManager;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class SpreadsheetManagerTest extends TestCase
{
    private SpreadsheetManager $manager;
    /** @var UserSpreadsheetRepository&MockObject */
    private UserSpreadsheetRepository $spreadsheetRepository;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var GoogleApiClientInterface&MockObject */
    private GoogleApiClientInterface $client;

    protected function setUp(): void
    {
        $this->spreadsheetRepository = $this->createMock(UserSpreadsheetRepository::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->client = $this->createMock(GoogleApiClientInterface::class);

        $this->manager = new SpreadsheetManager(
            $this->client,
            $this->spreadsheetRepository,
            $this->logger,
        );
    }

    public function testGetExpenseCategories(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $categories = [
            ['Продукты'],
            ['Транспорт'],
            [''],
            ['Развлечения'],
        ];

        $this->client->expects($this->once())
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!B28:B')
            ->willReturn($categories);

        $result = $this->manager->getExpenseCategories($spreadsheetId);

        $this->assertEquals(['Продукты', 'Транспорт', 'Развлечения'], $result);
    }

    public function testGetExpenseCategoriesWhenEmpty(): void
    {
        $spreadsheetId = 'test-spreadsheet';

        $this->client->expects($this->once())
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!B28:B')
            ->willReturn(null);

        $result = $this->manager->getExpenseCategories($spreadsheetId);

        $this->assertEquals([], $result);
    }

    public function testGetIncomeCategories(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $categories = [
            ['Зарплата'],
            ['Премия'],
            [''],
            ['Инвестиции'],
        ];

        $this->client->expects($this->once())
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!H28:H')
            ->willReturn($categories);

        $result = $this->manager->getIncomeCategories($spreadsheetId);

        $this->assertEquals(['Зарплата', 'Премия', 'Инвестиции'], $result);
    }

    public function testGetIncomeCategoriesWhenEmpty(): void
    {
        $spreadsheetId = 'test-spreadsheet';

        $this->client->expects($this->once())
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!H28:H')
            ->willReturn(null);

        $result = $this->manager->getIncomeCategories($spreadsheetId);

        $this->assertEquals([], $result);
    }

    public function testAddExpenseCategory(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $category = 'Новая категория';
        $existingCategories = [['Продукты'], ['Транспорт']];

        $this->client->expects($this->exactly(2))
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!B28:B')
            ->willReturn($existingCategories);

        $this->client->expects($this->once())
            ->method('updateValues')
            ->with(
                $spreadsheetId,
                'Сводка!B30:F30',
                [[$category, '', '', '', '']]
            );

        $this->manager->addExpenseCategory($spreadsheetId, $category);
    }

    public function testAddExpenseCategoryWhenEmpty(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $category = 'Новая категория';

        $this->client->expects($this->exactly(2))
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!B28:B')
            ->willReturn([]);

        $this->client->expects($this->once())
            ->method('updateValues')
            ->with(
                $spreadsheetId,
                'Сводка!B28:B',
                [[$category]]
            );

        $this->manager->addExpenseCategory($spreadsheetId, $category);
    }

    public function testAddExpenseCategoryWhenExists(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $category = 'Продукты';

        $this->client->expects($this->once())
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!B28:B')
            ->willReturn([['Продукты'], ['Транспорт']]);

        // Should not call updateValues since category already exists
        $this->client->expects($this->never())
            ->method('updateValues');

        $this->manager->addExpenseCategory($spreadsheetId, $category);
    }

    public function testAddIncomeCategory(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $category = 'Новая категория';
        $existingCategories = [['Зарплата'], ['Премия']];

        $this->client->expects($this->exactly(2))
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!H28:H')
            ->willReturn($existingCategories);

        $this->client->expects($this->once())
            ->method('updateValues')
            ->with(
                $spreadsheetId,
                'Сводка!H30:L30',
                [[$category, '', '', '', '']]
            );

        $this->manager->addIncomeCategory($spreadsheetId, $category);
    }

    public function testAddIncomeCategoryWhenEmpty(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $category = 'Новая категория';

        $this->client->expects($this->exactly(2))
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!H28:H')
            ->willReturn([]);

        $this->client->expects($this->once())
            ->method('updateValues')
            ->with(
                $spreadsheetId,
                'Сводка!H28:H',
                [[$category]]
            );

        $this->manager->addIncomeCategory($spreadsheetId, $category);
    }

    public function testAddIncomeCategoryWhenExists(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $category = 'Зарплата';

        $this->client->expects($this->once())
            ->method('getValues')
            ->with($spreadsheetId, 'Сводка!H28:H')
            ->willReturn([['Зарплата'], ['Премия']]);

        // Should not call updateValues since category already exists
        $this->client->expects($this->never())
            ->method('updateValues');

        $this->manager->addIncomeCategory($spreadsheetId, $category);
    }

    public function testHandleSpreadsheetIdWithUrl(): void
    {
        $url = 'https://docs.google.com/spreadsheets/d/abc123/edit#gid=0';
        $expectedId = 'abc123';

        $this->client->expects($this->once())
            ->method('validateSpreadsheetAccess')
            ->with($expectedId)
            ->willReturn(true);

        $result = $this->manager->handleSpreadsheetId($url);
        $this->assertEquals($expectedId, $result);
    }

    public function testHandleSpreadsheetIdWithInvalidUrl(): void
    {
        $url = 'https://docs.google.com/spreadsheets/d/';

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Неверный формат ссылки. Пожалуйста, убедитесь, что вы скопировали полную ссылку на таблицу.');

        $this->manager->handleSpreadsheetId($url);
    }

    public function testHandleSpreadsheetIdWithoutAccess(): void
    {
        $id = 'abc123';

        $this->client->expects($this->once())
            ->method('validateSpreadsheetAccess')
            ->with($id)
            ->willReturn(false);

        $this->client->expects($this->once())
            ->method('getSharingInstructions')
            ->with($id)
            ->willReturn('Please share the spreadsheet');

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Please share the spreadsheet');

        $this->manager->handleSpreadsheetId($id);
    }

    public function testFindSpreadsheetByDate(): void
    {
        $user = new User();
        $date = new \DateTime();
        $spreadsheet = new UserSpreadsheet();

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByDate')
            ->with($user, $date)
            ->willReturn($spreadsheet);

        $result = $this->manager->findSpreadsheetByDate($user, $date);
        $this->assertSame($spreadsheet, $result);
    }

    public function testFindLatestSpreadsheet(): void
    {
        $user = new User();
        $spreadsheet = new UserSpreadsheet();

        $this->spreadsheetRepository->expects($this->once())
            ->method('findLatest')
            ->with($user)
            ->willReturn($spreadsheet);

        $result = $this->manager->findLatestSpreadsheet($user);
        $this->assertSame($spreadsheet, $result);
    }

    public function testAddSpreadsheet(): void
    {
        $user = new User();
        $spreadsheetId = 'test-spreadsheet';
        $month = 1;
        $year = 2024;
        $title = 'Test Budget';

        $this->client->expects($this->once())
            ->method('getSpreadsheetTitle')
            ->with($spreadsheetId)
            ->willReturn($title);

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByMonthAndYear')
            ->with($user, $month, $year)
            ->willReturn(null);

        $this->spreadsheetRepository->expects($this->once())
            ->method('save')
            ->with($this->callback(function (UserSpreadsheet $spreadsheet) use ($user, $spreadsheetId, $month, $year, $title) {
                return $spreadsheet->getUser() === $user
                    && $spreadsheet->getSpreadsheetId() === $spreadsheetId
                    && $spreadsheet->getMonth() === $month
                    && $spreadsheet->getYear() === $year
                    && $spreadsheet->getTitle() === $title;
            }), true);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Spreadsheet added for user', [
                'user_id' => null,
                'spreadsheet_id' => $spreadsheetId,
                'title' => $title,
                'month' => $month,
                'year' => $year,
            ]);

        $this->manager->addSpreadsheet($user, $spreadsheetId, $month, $year);
    }

    public function testAddSpreadsheetFailsWhenTitleNotFound(): void
    {
        $user = new User();
        $spreadsheetId = 'test-spreadsheet';
        $month = 1;
        $year = 2024;

        $this->client->expects($this->once())
            ->method('getSpreadsheetTitle')
            ->with($spreadsheetId)
            ->willReturn(null);

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Failed to get spreadsheet title');

        $this->manager->addSpreadsheet($user, $spreadsheetId, $month, $year);
    }

    public function testAddSpreadsheetFailsWhenSpreadsheetExists(): void
    {
        $user = new User();
        $spreadsheetId = 'test-spreadsheet';
        $month = 1;
        $year = 2024;
        $title = 'Test Budget';

        $this->client->expects($this->once())
            ->method('getSpreadsheetTitle')
            ->with($spreadsheetId)
            ->willReturn($title);

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByMonthAndYear')
            ->with($user, $month, $year)
            ->willReturn(new UserSpreadsheet());

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Таблица для этого месяца и года уже существует');

        $this->manager->addSpreadsheet($user, $spreadsheetId, $month, $year);
    }

    public function testRemoveSpreadsheet(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $month = 1;
        $year = 2024;
        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setSpreadsheetId('test-spreadsheet');

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByMonthAndYear')
            ->with($user, $month, $year)
            ->willReturn($spreadsheet);

        $this->client->expects($this->once())
            ->method('validateSpreadsheetAccess')
            ->with('test-spreadsheet')
            ->willReturn(true);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Removing spreadsheet {spreadsheet_id} for user {telegram_id}', [
                'spreadsheet_id' => 'test-spreadsheet',
                'telegram_id' => 123456,
                'month' => $month,
                'year' => $year,
            ]);

        $this->spreadsheetRepository->expects($this->once())
            ->method('remove')
            ->with($spreadsheet, true);

        $this->manager->removeSpreadsheet($user, $month, $year);
    }

    public function testRemoveSpreadsheetFailsWhenNotFound(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $month = 1;
        $year = 2024;

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByMonthAndYear')
            ->with($user, $month, $year)
            ->willReturn(null);

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Таблица за Январь 2024 не найдена');

        $this->manager->removeSpreadsheet($user, $month, $year);
    }

    public function testRemoveSpreadsheetFailsWhenNoAccess(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $month = 1;
        $year = 2024;
        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setSpreadsheetId('test-spreadsheet');

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByMonthAndYear')
            ->with($user, $month, $year)
            ->willReturn($spreadsheet);

        $this->client->expects($this->once())
            ->method('validateSpreadsheetAccess')
            ->with('test-spreadsheet')
            ->willReturn(false);

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Не удалось получить доступ к таблице');

        $this->manager->removeSpreadsheet($user, $month, $year);
    }

    public function testGetSpreadsheetsList(): void
    {
        $user = new User();
        $spreadsheet1 = new UserSpreadsheet();
        $spreadsheet1->setMonth(1);
        $spreadsheet1->setYear(2024);
        $spreadsheet1->setSpreadsheetId('spreadsheet1');

        $spreadsheet2 = new UserSpreadsheet();
        $spreadsheet2->setMonth(2);
        $spreadsheet2->setYear(2024);
        $spreadsheet2->setSpreadsheetId('spreadsheet2');

        $this->spreadsheetRepository->expects($this->once())
            ->method('findBy')
            ->with(['user' => $user], ['year' => 'DESC', 'month' => 'DESC'])
            ->willReturn([$spreadsheet1, $spreadsheet2]);

        $result = $this->manager->getSpreadsheetsList($user);

        $this->assertEquals([
            [
                'month' => 'Январь',
                'year' => 2024,
                'url' => 'https://docs.google.com/spreadsheets/d/spreadsheet1',
            ],
            [
                'month' => 'Февраль',
                'year' => 2024,
                'url' => 'https://docs.google.com/spreadsheets/d/spreadsheet2',
            ],
        ], $result);
    }

    public function testGetSpreadsheetsListSkipsInvalidSpreadsheets(): void
    {
        $user = new User();
        $spreadsheet1 = new UserSpreadsheet();
        $spreadsheet1->setMonth(1);
        $spreadsheet1->setYear(2024);
        $spreadsheet1->setSpreadsheetId('spreadsheet1');

        $spreadsheet2 = new UserSpreadsheet();
        // Missing month, year, and spreadsheetId

        $this->spreadsheetRepository->expects($this->once())
            ->method('findBy')
            ->with(['user' => $user], ['year' => 'DESC', 'month' => 'DESC'])
            ->willReturn([$spreadsheet1, $spreadsheet2]);

        $result = $this->manager->getSpreadsheetsList($user);

        $this->assertEquals([
            [
                'month' => 'Январь',
                'year' => 2024,
                'url' => 'https://docs.google.com/spreadsheets/d/spreadsheet1',
            ],
        ], $result);
    }
}
