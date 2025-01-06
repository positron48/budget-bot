<?php

namespace App\Service;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use App\Service\Google\GoogleApiClientInterface;
use App\Service\Google\SpreadsheetManager;
use App\Service\Google\TransactionRecorder;
use Psr\Log\LoggerInterface;

class GoogleSheetsService
{
    protected SpreadsheetManager $spreadsheetManager;
    protected TransactionRecorder $transactionRecorder;
    protected CategoryService $categoryService;

    public function __construct(
        GoogleApiClientInterface $client,
        UserSpreadsheetRepository $spreadsheetRepository,
        LoggerInterface $logger,
        CategoryService $categoryService,
    ) {
        $this->spreadsheetManager = new SpreadsheetManager($client, $spreadsheetRepository, $logger);
        $this->transactionRecorder = new TransactionRecorder($client, $logger);
        $this->categoryService = $categoryService;
    }

    public function addExpense(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $this->transactionRecorder->addExpense($spreadsheetId, $date, $amount, $description, $category);
    }

    public function addIncome(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $this->transactionRecorder->addIncome($spreadsheetId, $date, $amount, $description, $category);
    }

    public function addSpreadsheet(User $user, string $spreadsheetId, int $month, int $year): void
    {
        $this->spreadsheetManager->addSpreadsheet($user, $spreadsheetId, $month, $year);

        // Import categories from the spreadsheet
        $this->syncCategories($user, $spreadsheetId);
    }

    public function removeSpreadsheet(User $user, int $month, int $year): void
    {
        $this->spreadsheetManager->removeSpreadsheet($user, $month, $year);
    }

    public function findSpreadsheetByDate(User $user, \DateTime $date): ?UserSpreadsheet
    {
        return $this->spreadsheetManager->findSpreadsheetByDate($user, $date);
    }

    public function findLatestSpreadsheet(User $user): ?UserSpreadsheet
    {
        return $this->spreadsheetManager->findLatestSpreadsheet($user);
    }

    /**
     * @return array<int, array{month: string, year: int, url: string}>
     */
    public function getSpreadsheetsList(User $user): array
    {
        return $this->spreadsheetManager->getSpreadsheetsList($user);
    }

    public function handleSpreadsheetId(string $input): string
    {
        return $this->spreadsheetManager->handleSpreadsheetId($input);
    }

    /**
     * @return array{
     *     added_to_db: array{
     *         expense: array<string>,
     *         income: array<string>
     *     },
     *     added_to_sheet: array{
     *         expense: array<string>,
     *         income: array<string>
     *     }
     * }
     */
    public function syncCategories(User $user, string $spreadsheetId): array
    {
        $changes = [
            'added_to_db' => [
                'expense' => [],
                'income' => [],
            ],
            'added_to_sheet' => [
                'expense' => [],
                'income' => [],
            ],
        ];

        // Get categories from spreadsheet
        $expenseCategories = $this->spreadsheetManager->getExpenseCategories($spreadsheetId);
        $incomeCategories = $this->spreadsheetManager->getIncomeCategories($spreadsheetId);

        // Get categories from database
        $dbExpenseCategories = $this->categoryService->getCategories(false, $user);
        $dbIncomeCategories = $this->categoryService->getCategories(true, $user);

        // Add missing categories to database
        foreach ($expenseCategories as $category) {
            if (!in_array($category, $dbExpenseCategories, true)) {
                $this->categoryService->addUserCategory($user, $category, false);
                $changes['added_to_db']['expense'][] = $category;
            }
        }

        foreach ($incomeCategories as $category) {
            if (!in_array($category, $dbIncomeCategories, true)) {
                $this->categoryService->addUserCategory($user, $category, true);
                $changes['added_to_db']['income'][] = $category;
            }
        }

        // Add missing categories to spreadsheet
        foreach ($dbExpenseCategories as $category) {
            if (!in_array($category, $expenseCategories, true)) {
                $this->spreadsheetManager->addExpenseCategory($spreadsheetId, $category);
                $changes['added_to_sheet']['expense'][] = $category;
            }
        }

        foreach ($dbIncomeCategories as $category) {
            if (!in_array($category, $incomeCategories, true)) {
                $this->spreadsheetManager->addIncomeCategory($spreadsheetId, $category);
                $changes['added_to_sheet']['income'][] = $category;
            }
        }

        return $changes;
    }
}
