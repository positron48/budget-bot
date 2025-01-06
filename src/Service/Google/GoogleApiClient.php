<?php

namespace App\Service\Google;

use Google\Service\Drive;
use Google\Service\Sheets;
use Google\Service\Sheets\ValueRange;
use Psr\Log\LoggerInterface;

class GoogleApiClient implements GoogleApiClientInterface
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

            if ($response instanceof ValueRange) {
                return $response->getValues();
            }

            return null;
        } catch (\Exception $e) {
            $this->logger->error('Failed to get values from spreadsheet: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
                'spreadsheetId' => $spreadsheetId,
                'range' => $range,
            ]);

            return null;
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
                ['valueInputOption' => 'USER_ENTERED']
            );
        } catch (\Exception $e) {
            $this->logger->error('Failed to update values in spreadsheet: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
                'spreadsheetId' => $spreadsheetId,
                'range' => $range,
            ]);
            throw $e;
        }
    }

    public function validateSpreadsheetAccess(string $spreadsheetId): bool
    {
        try {
            $this->sheetsService->spreadsheets->get($spreadsheetId);

            return true;
        } catch (\Exception $e) {
            $this->logger->info('Failed to validate spreadsheet access: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
                'spreadsheetId' => $spreadsheetId,
            ]);

            return false;
        }
    }

    public function getSpreadsheetTitle(string $spreadsheetId): ?string
    {
        try {
            $spreadsheet = $this->sheetsService->spreadsheets->get($spreadsheetId);

            return $spreadsheet->getProperties()->getTitle();
        } catch (\Exception $e) {
            $this->logger->error('Failed to get spreadsheet title: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
                'spreadsheetId' => $spreadsheetId,
            ]);

            return null;
        }
    }

    public function cloneSpreadsheet(string $sourceId, string $newTitle): string
    {
        try {
            $copy = $this->driveService->files->copy($sourceId, [
                'name' => $newTitle,
            ]);

            $this->driveService->permissions->create($copy->getId(), [
                'role' => 'writer',
                'type' => 'user',
                'emailAddress' => $this->serviceAccountEmail,
            ]);

            return $copy->getId();
        } catch (\Exception $e) {
            $this->logger->error('Failed to clone spreadsheet: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
                'sourceId' => $sourceId,
                'newTitle' => $newTitle,
            ]);
            throw $e;
        }
    }

    public function getSharingInstructions(string $spreadsheetId): string
    {
        return sprintf(
            'Для работы с таблицей предоставьте доступ на редактирование для %s',
            $this->serviceAccountEmail
        );
    }
}
