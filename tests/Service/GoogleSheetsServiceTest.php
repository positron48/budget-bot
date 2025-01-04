<?php

namespace Tests\Service;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use App\Service\GoogleSheetsService;
use Google\Service\Drive;
use Google\Service\Sheets;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class GoogleSheetsServiceTest extends TestCase
{
    /** @var UserSpreadsheetRepository&MockObject */
    private UserSpreadsheetRepository $spreadsheetRepository;

    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;

    /** @var Drive&MockObject */
    private Drive $driveService;

    /** @var Sheets&MockObject */
    private Sheets $sheetsService;

    private GoogleSheetsService $service;

    protected function setUp(): void
    {
        $this->spreadsheetRepository = $this->createMock(UserSpreadsheetRepository::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->driveService = $this->createMock(Drive::class);
        $this->sheetsService = $this->createMock(Sheets::class);

        // Create test credentials file
        $credentialsDir = dirname(__DIR__, 2).'/var/test';
        $credentialsFile = $credentialsDir.'/credentials.json';

        if (!file_exists($credentialsDir)) {
            mkdir($credentialsDir, 0777, true);
        }
        file_put_contents($credentialsFile, json_encode([
            'type' => 'service_account',
            'project_id' => 'test-project',
            'private_key_id' => 'test-key-id',
            'private_key' => '-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n',
            'client_email' => 'test@example.com',
            'client_id' => 'test-client-id',
            'auth_uri' => 'https://accounts.google.com/o/oauth2/auth',
            'token_uri' => 'https://oauth2.googleapis.com/token',
            'auth_provider_x509_cert_url' => 'https://www.googleapis.com/oauth2/v1/certs',
            'client_x509_cert_url' => 'https://www.googleapis.com/robot/v1/metadata/x509/test%40example.com',
        ]));

        // Create service with mocked dependencies
        $this->service = new GoogleSheetsService(
            $credentialsFile,
            'test@example.com',
            $this->logger,
            $this->spreadsheetRepository
        );

        // Inject mocked services
        $reflection = new \ReflectionClass($this->service);

        $driveProperty = $reflection->getProperty('driveService');
        $driveProperty->setAccessible(true);
        $driveProperty->setValue($this->service, $this->driveService);

        $sheetsProperty = $reflection->getProperty('sheetsService');
        $sheetsProperty->setAccessible(true);
        $sheetsProperty->setValue($this->service, $this->sheetsService);
    }

    protected function tearDown(): void
    {
        parent::tearDown();

        // Clean up test credentials file
        $credentialsDir = dirname(__DIR__, 2).'/var/test';
        $credentialsFile = $credentialsDir.'/credentials.json';

        if (file_exists($credentialsFile)) {
            unlink($credentialsFile);
        }
        if (file_exists($credentialsDir)) {
            rmdir($credentialsDir);
        }
    }

    public function testGetSpreadsheetUrl(): void
    {
        $spreadsheetId = 'test-id';
        $expectedUrl = 'https://docs.google.com/spreadsheets/d/test-id';

        $this->assertEquals($expectedUrl, $this->service->getSpreadsheetUrl($spreadsheetId));
    }

    public function testRemoveSpreadsheet(): void
    {
        $user = new User();
        $user->setTelegramId(123);

        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setSpreadsheetId('test-id');
        $spreadsheet->setMonth(1); // January
        $spreadsheet->setYear(2024);
        $spreadsheet->setUser($user);

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByMonthAndYear')
            ->with($user, 1, 2024)
            ->willReturn($spreadsheet);

        $this->spreadsheetRepository->expects($this->once())
            ->method('remove')
            ->with($spreadsheet);

        $this->service->removeSpreadsheet($user, 1, 2024);
    }

    public function testRemoveSpreadsheetNotFound(): void
    {
        $user = new User();
        $user->setTelegramId(123);

        $this->spreadsheetRepository->expects($this->once())
            ->method('findByMonthAndYear')
            ->with($user, 1, 2024)
            ->willReturn(null);

        $this->spreadsheetRepository->expects($this->never())
            ->method('remove');

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Таблица за Январь 2024 не найдена');

        $this->service->removeSpreadsheet($user, 1, 2024);
    }

    public function testGetSpreadsheetsList(): void
    {
        $user = new User();
        $user->setTelegramId(123);

        $spreadsheet1 = new UserSpreadsheet();
        $spreadsheet1->setSpreadsheetId('test-id-1');
        $spreadsheet1->setMonth(1); // January
        $spreadsheet1->setYear(2024);

        $spreadsheet2 = new UserSpreadsheet();
        $spreadsheet2->setSpreadsheetId('test-id-2');
        $spreadsheet2->setMonth(2); // February
        $spreadsheet2->setYear(2024);

        $this->spreadsheetRepository->expects($this->once())
            ->method('findBy')
            ->with(['user' => $user], ['year' => 'DESC', 'month' => 'DESC'])
            ->willReturn([$spreadsheet1, $spreadsheet2]);

        $expected = [
            [
                'month' => 'Январь',
                'year' => 2024,
                'url' => 'https://docs.google.com/spreadsheets/d/test-id-1',
            ],
            [
                'month' => 'Февраль',
                'year' => 2024,
                'url' => 'https://docs.google.com/spreadsheets/d/test-id-2',
            ],
        ];

        $this->assertEquals($expected, $this->service->getSpreadsheetsList($user));
    }
}
