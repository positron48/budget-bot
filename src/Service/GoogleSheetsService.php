<?php

namespace App\Service;

use Google\Client;
use Google\Service\Sheets;
use Google\Service\Sheets\ValueRange;

class GoogleSheetsService
{
    private Sheets $sheetsService;
    private const FIRST_DATA_ROW = 5;
    private const DATE_COLUMN = 'B';
    private const AMOUNT_COLUMN = 'C';
    private const DESCRIPTION_COLUMN = 'D';
    private const CATEGORY_COLUMN = 'E';
    private const INCOME_DATE_COLUMN = 'G';
    private const INCOME_AMOUNT_COLUMN = 'H';
    private const INCOME_DESCRIPTION_COLUMN = 'I';
    private const INCOME_CATEGORY_COLUMN = 'J';

    public function __construct(string $credentialsPath)
    {
        $client = new Client();
        $client->setAuthConfig($credentialsPath);
        $client->setScopes([Sheets::SPREADSHEETS]);
        $this->sheetsService = new Sheets($client);
    }

    public function addExpense(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category
    ): void {
        $range = 'Транзакции!B:E';
        $row = $this->findFirstEmptyRow($spreadsheetId, $range);
        
        $values = [
            [
                $date,
                $amount,
                $description,
                $category
            ]
        ];

        $body = new ValueRange([
            'values' => $values
        ]);

        $this->sheetsService->spreadsheets_values->update(
            $spreadsheetId,
            "Транзакции!B{$row}:E{$row}",
            $body,
            ['valueInputOption' => 'USER_ENTERED']
        );
    }

    public function addIncome(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category
    ): void {
        $range = 'Транзакции!G:J';
        $row = $this->findFirstEmptyRow($spreadsheetId, $range);
        
        $values = [
            [
                $date,
                $amount,
                $description,
                $category
            ]
        ];

        $body = new ValueRange([
            'values' => $values
        ]);

        $this->sheetsService->spreadsheets_values->update(
            $spreadsheetId,
            "Транзакции!G{$row}:J{$row}",
            $body,
            ['valueInputOption' => 'USER_ENTERED']
        );
    }

    private function findFirstEmptyRow(string $spreadsheetId, string $range): int
    {
        $response = $this->sheetsService->spreadsheets_values->get($spreadsheetId, $range);
        $values = $response->getValues();
        
        return count($values ?? []) + self::FIRST_DATA_ROW;
    }

    public function getSpreadsheetsList(): array
    {
        // This method should be implemented based on how you want to manage spreadsheet access
        // You might want to store spreadsheet IDs in a configuration file or database
        return [];
    }
} 