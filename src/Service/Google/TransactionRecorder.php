<?php

namespace App\Service\Google;

use Psr\Log\LoggerInterface;

class TransactionRecorder
{
    private const EXPENSE_RANGE = 'Транзакции!B:E';
    private const INCOME_RANGE = 'Транзакции!G:J';

    private GoogleSheetsClient $client;
    private LoggerInterface $logger;

    public function __construct(
        GoogleSheetsClient $client,
        LoggerInterface $logger,
    ) {
        $this->client = $client;
        $this->logger = $logger;
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

        $values = [
            [
                $date,
                $amount,
                $description,
                $category,
            ],
        ];

        $this->client->updateValues($spreadsheetId, self::EXPENSE_RANGE, $values);

        $this->logger->info('Expense added successfully', [
            'spreadsheet_id' => $spreadsheetId,
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

        $values = [
            [
                $date,
                $amount,
                $description,
                $category,
            ],
        ];

        $this->client->updateValues($spreadsheetId, self::INCOME_RANGE, $values);

        $this->logger->info('Income added successfully', [
            'spreadsheet_id' => $spreadsheetId,
        ]);
    }
}
