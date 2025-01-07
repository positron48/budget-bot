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
}
