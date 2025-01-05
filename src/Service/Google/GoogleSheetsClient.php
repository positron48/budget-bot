<?php

namespace App\Service\Google;

use Google\Service\Drive;
use Google\Service\Sheets;
use Google\Service\Sheets\ValueRange;
use Psr\Log\LoggerInterface;

class GoogleSheetsClient
{
    private Sheets $sheetsService;
    private Drive $driveService;
    private string $serviceAccountEmail;
    private LoggerInterface $logger;

    public function __construct(
        string $credentialsPath,
        string $serviceAccountEmail,
        LoggerInterface $logger,
    ) {
        $this->serviceAccountEmail = $serviceAccountEmail;
        $this->logger = $logger;

        try {
            $client = new \Google\Client();
            $client->setAuthConfig($credentialsPath);
            $client->setScopes([
                Sheets::SPREADSHEETS,
                Drive::DRIVE_READONLY,
            ]);

            $this->sheetsService = new Sheets($client);
            $this->driveService = new Drive($client);
        } catch (\Exception $e) {
            $this->logger->error('Failed to initialize Google services: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    /**
     * @return array<int, array<int, string|float>>|null
     */
    public function getValues(string $spreadsheetId, string $range): ?array
    {
        try {
            $response = $this->sheetsService->spreadsheets_values->get($spreadsheetId, $range);

            return $response->getValues();
        } catch (\Exception $e) {
            $this->logger->error('Failed to get values: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
                'range' => $range,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    /**
     * @param array<int, array<int, string|float>> $values
     */
    public function updateValues(string $spreadsheetId, string $range, array $values): void
    {
        try {
            $body = new ValueRange([
                'values' => $values,
            ]);

            $this->sheetsService->spreadsheets_values->update(
                $spreadsheetId,
                $range,
                $body,
                ['valueInputOption' => 'RAW']
            );
        } catch (\Exception $e) {
            $this->logger->error('Failed to update values: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
                'range' => $range,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    public function validateSpreadsheetAccess(string $spreadsheetId): bool
    {
        try {
            $this->sheetsService->spreadsheets->get($spreadsheetId);
            $this->logger->info('Spreadsheet access validated', [
                'spreadsheet_id' => $spreadsheetId,
            ]);

            return true;
        } catch (\Exception $e) {
            $this->logger->warning('Failed to validate spreadsheet access: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
            ]);

            return false;
        }
    }

    public function getSpreadsheetTitle(string $spreadsheetId): ?string
    {
        try {
            $spreadsheet = $this->sheetsService->spreadsheets->get($spreadsheetId);
            $title = $spreadsheet->getProperties()->getTitle();
            $this->logger->info('Retrieved spreadsheet title', [
                'spreadsheet_id' => $spreadsheetId,
                'title' => $title,
            ]);

            return $title;
        } catch (\Exception $e) {
            $this->logger->error('Failed to get spreadsheet title: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
            ]);

            return null;
        }
    }

    public function cloneSpreadsheet(string $sourceId, string $newTitle): string
    {
        try {
            $copy = $this->driveService->files->copy(
                $sourceId,
                new Drive\DriveFile(['name' => $newTitle])
            );

            $this->logger->info('Spreadsheet cloned', [
                'source_id' => $sourceId,
                'new_id' => $copy->getId(),
                'new_title' => $newTitle,
            ]);

            return $copy->getId();
        } catch (\Exception $e) {
            $this->logger->error('Failed to clone spreadsheet: {error}', [
                'error' => $e->getMessage(),
                'source_id' => $sourceId,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    public function getSharingInstructions(string $spreadsheetId): string
    {
        $this->logger->info('Generated sharing instructions', [
            'spreadsheet_id' => $spreadsheetId,
            'service_account' => $this->serviceAccountEmail,
        ]);

        return "Для работы с таблицей нужно предоставить доступ сервисному аккаунту:\n\n".
               "1. Откройте таблицу\n".
               "2. Нажмите кнопку \"Настройки доступа\" (или \"Share\")\n".
               "3. В поле \"Добавить пользователей или группы\" введите:\n".
               "{$this->serviceAccountEmail}\n".
               "4. Выберите роль \"Редактор\"\n".
               "5. Нажмите \"Готово\"\n\n".
               'После этого отправьте команду /add еще раз';
    }
}
