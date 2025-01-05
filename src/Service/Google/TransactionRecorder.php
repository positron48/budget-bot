<?php

namespace App\Service\Google;

use Psr\Log\LoggerInterface;

class TransactionRecorder
{
    private const EXPENSE_RANGE = 'Транзакции!B5:E';
    private const INCOME_RANGE = 'Транзакции!G5:J';

    private GoogleSheetsClient $client;
    private LoggerInterface $logger;

    public function __construct(
        GoogleSheetsClient $client,
        LoggerInterface $logger,
    ) {
        $this->client = $client;
        $this->logger = $logger;
    }

    private function findFirstEmptyRow(string $spreadsheetId, string $range): int
    {
        $values = $this->client->getValues($spreadsheetId, $range);
        $row = 5; // Start from row 5

        if (!empty($values)) {
            $row = count($values) + 5;
        }

        return $row;
    }

    public function addExpense(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $this->logger->info('Adding expense', [
            'spreadsheet_id' => $spreadsheetId,
            'date' => $date,
            'amount' => $amount,
            'description' => $description,
            'category' => $category,
        ]);

        $row = $this->findFirstEmptyRow($spreadsheetId, self::EXPENSE_RANGE);
        $range = sprintf('Транзакции!B%d:E%d', $row, $row);

        $values = [
            [
                $date,
                $amount,
                $description,
                $category,
            ],
        ];

        $this->client->updateValues($spreadsheetId, $range, $values);

        $this->logger->info('Expense added successfully', [
            'spreadsheet_id' => $spreadsheetId,
            'row' => $row,
        ]);
    }

    public function addIncome(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $this->logger->info('Adding income', [
            'spreadsheet_id' => $spreadsheetId,
            'date' => $date,
            'amount' => $amount,
            'description' => $description,
            'category' => $category,
        ]);

        $row = $this->findFirstEmptyRow($spreadsheetId, self::INCOME_RANGE);
        $range = sprintf('Транзакции!G%d:J%d', $row, $row);

        $values = [
            [
                $date,
                $amount,
                $description,
                $category,
            ],
        ];

        $this->client->updateValues($spreadsheetId, $range, $values);

        $this->logger->info('Income added successfully', [
            'spreadsheet_id' => $spreadsheetId,
            'row' => $row,
        ]);
    }
}
