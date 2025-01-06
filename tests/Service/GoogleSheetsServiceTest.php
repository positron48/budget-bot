<?php

namespace App\Tests\Service;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use App\Service\CategoryService;
use App\Service\Google\GoogleSheetsClient;
use App\Service\Google\SpreadsheetManager;
use App\Service\Google\TransactionRecorder;
use App\Service\GoogleSheetsService;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class GoogleSheetsServiceTest extends TestCase
{
    private const MOCK_CREDENTIALS_PATH = __DIR__.'/../config/google-credentials.json';
    private const SERVICE_ACCOUNT_EMAIL = 'test@example.com';

    private GoogleSheetsService $service;
    /** @var UserSpreadsheetRepository&MockObject */
    private UserSpreadsheetRepository $spreadsheetRepository;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var GoogleSheetsClient&MockObject */
    private GoogleSheetsClient $client;
    /** @var CategoryService&MockObject */
    private CategoryService $categoryService;

    protected function setUp(): void
    {
        $this->spreadsheetRepository = $this->createMock(UserSpreadsheetRepository::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->client = $this->createMock(GoogleSheetsClient::class);
        $this->categoryService = $this->createMock(CategoryService::class);

        $this->service = new GoogleSheetsService(
            self::MOCK_CREDENTIALS_PATH,
            self::SERVICE_ACCOUNT_EMAIL,
            $this->logger,
            $this->spreadsheetRepository,
            $this->categoryService,
        );

        // Replace service dependencies with mocks
        $reflection = new \ReflectionClass($this->service);
        $spreadsheetManagerProperty = $reflection->getProperty('spreadsheetManager');
        $spreadsheetManagerProperty->setValue($this->service, new SpreadsheetManager(
            $this->client,
            $this->spreadsheetRepository,
            $this->logger,
        ));
        $transactionRecorderProperty = $reflection->getProperty('transactionRecorder');
        $transactionRecorderProperty->setValue($this->service, new TransactionRecorder(
            $this->client,
            $this->logger,
        ));
    }

    public function testGetSpreadsheetsList(): void
    {
        $user = new User();
        $user->setTelegramId(123456);

        $spreadsheet1 = new UserSpreadsheet();
        $spreadsheet1->setUser($user)
            ->setSpreadsheetId('spreadsheet1')
            ->setTitle('Test Spreadsheet 1')
            ->setMonth(1)
            ->setYear(2024);

        $spreadsheet2 = new UserSpreadsheet();
        $spreadsheet2->setUser($user)
            ->setSpreadsheetId('spreadsheet2')
            ->setTitle('Test Spreadsheet 2')
            ->setMonth(2)
            ->setYear(2024);

        $this->spreadsheetRepository->method('findBy')
            ->with(['user' => $user], ['year' => 'DESC', 'month' => 'DESC'])
            ->willReturn([$spreadsheet1, $spreadsheet2]);

        $result = $this->service->getSpreadsheetsList($user);

        $this->assertCount(2, $result);
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

    public function testHandleSpreadsheetId(): void
    {
        $url = 'https://docs.google.com/spreadsheets/d/abc123/edit#gid=0';
        $this->client->method('validateSpreadsheetAccess')
            ->with('abc123')
            ->willReturn(true);

        $id = $this->service->handleSpreadsheetId($url);
        $this->assertEquals('abc123', $id);
    }

    public function testAddExpense(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $date = '2024-01-05';
        $amount = 1000.0;
        $description = 'Test expense';
        $category = 'Test category';

        $this->client->expects($this->once())
            ->method('getValues')
            ->willReturn([]);

        $this->client->expects($this->once())
            ->method('updateValues')
            ->with(
                $spreadsheetId,
                'Транзакции!B5:E5',
                [[$date, $amount, $description, $category]]
            );

        $this->service->addExpense($spreadsheetId, $date, $amount, $description, $category);
    }

    public function testAddIncome(): void
    {
        $spreadsheetId = 'test-spreadsheet';
        $date = '2024-01-05';
        $amount = 5000.0;
        $description = 'Test income';
        $category = 'Salary';

        $this->client->expects($this->once())
            ->method('getValues')
            ->willReturn([]);

        $this->client->expects($this->once())
            ->method('updateValues')
            ->with(
                $spreadsheetId,
                'Транзакции!G5:J5',
                [[$date, $amount, $description, $category]]
            );

        $this->service->addIncome($spreadsheetId, $date, $amount, $description, $category);
    }

    public function testAddSpreadsheet(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $spreadsheetId = 'test-spreadsheet';
        $month = 1;
        $year = 2024;
        $title = 'Test Spreadsheet';

        $this->client->method('getSpreadsheetTitle')
            ->with($spreadsheetId)
            ->willReturn($title);

        $this->spreadsheetRepository->method('findByMonthAndYear')
            ->with($user, $month, $year)
            ->willReturn(null);

        $this->spreadsheetRepository->expects($this->once())
            ->method('save')
            ->with(
                $this->callback(function (UserSpreadsheet $spreadsheet) use ($user, $spreadsheetId, $month, $year, $title) {
                    return $spreadsheet->getUser() === $user
                        && $spreadsheet->getSpreadsheetId() === $spreadsheetId
                        && $spreadsheet->getMonth() === $month
                        && $spreadsheet->getYear() === $year
                        && $spreadsheet->getTitle() === $title;
                }),
                true
            );

        $this->service->addSpreadsheet($user, $spreadsheetId, $month, $year);
    }

    public function testRemoveSpreadsheet(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $month = 1;
        $year = 2024;

        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setUser($user)
            ->setSpreadsheetId('test-spreadsheet')
            ->setTitle('Test Spreadsheet')
            ->setMonth($month)
            ->setYear($year);

        $this->spreadsheetRepository->method('findByMonthAndYear')
            ->with($user, $month, $year)
            ->willReturn($spreadsheet);

        $this->spreadsheetRepository->expects($this->once())
            ->method('remove')
            ->with($spreadsheet, true);

        $this->service->removeSpreadsheet($user, $month, $year);
    }
}
