<?php

namespace App\Service;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use App\Service\Google\GoogleSheetsClient;
use App\Service\Google\SpreadsheetManager;
use App\Service\Google\TransactionRecorder;
use Psr\Log\LoggerInterface;

class GoogleSheetsService
{
    protected SpreadsheetManager $spreadsheetManager;
    protected TransactionRecorder $transactionRecorder;

    public function __construct(
        string $credentialsPath,
        string $serviceAccountEmail,
        LoggerInterface $logger,
        UserSpreadsheetRepository $spreadsheetRepository,
    ) {
        $client = new GoogleSheetsClient($credentialsPath, $serviceAccountEmail, $logger);
        $this->spreadsheetManager = new SpreadsheetManager($client, $spreadsheetRepository, $logger);
        $this->transactionRecorder = new TransactionRecorder($client, $logger);
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
}
