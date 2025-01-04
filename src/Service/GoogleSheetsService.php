<?php

namespace App\Service;

use Google\Service\Sheets;
use Google\Service\Sheets\ValueRange;

class GoogleSheetsService
{
    private const FIRST_DATA_ROW = 2;

    private Sheets $sheetsService;

    public function __construct(string $credentialsPath)
    {
        $client = new \Google\Client();
        $client->setAuthConfig($credentialsPath);
        $client->setScopes([Sheets::SPREADSHEETS]);

        $this->sheetsService = new Sheets($client);
    }

    public function addExpense(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $range = 'Транзакции!B:E';
        $row = $this->findFirstEmptyRow($spreadsheetId, $range);

        $values = [
            [
                $date,
                $amount,
                $description,
                $category,
            ],
        ];

        $body = new ValueRange([
            'values' => $values,
        ]);

        $this->sheetsService->spreadsheets_values->update(
            $spreadsheetId,
            $range,
            $body,
            ['valueInputOption' => 'RAW']
        );
    }

    public function addIncome(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $range = 'Транзакции!G:J';
        $row = $this->findFirstEmptyRow($spreadsheetId, $range);

        $values = [
            [
                $date,
                $amount,
                $description,
                $category,
            ],
        ];

        $body = new ValueRange([
            'values' => $values,
        ]);

        $this->sheetsService->spreadsheets_values->update(
            $spreadsheetId,
            $range,
            $body,
            ['valueInputOption' => 'RAW']
        );
    }

    private function findFirstEmptyRow(string $spreadsheetId, string $range): int
    {
        $response = $this->sheetsService->spreadsheets_values->get($spreadsheetId, $range);
        $values = $response->getValues();

        return count($values ?? []) + self::FIRST_DATA_ROW;
    }

    /**
     * @return string[]
     */
    public function getSpreadsheetsList(): array
    {
        // You might want to store spreadsheet IDs in a configuration file or database
        return [];
    }
}
