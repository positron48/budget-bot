<?php

namespace App\Tests\Mock;

use App\Service\Google\GoogleApiClientInterface;

class TestGoogleApiClient implements GoogleApiClientInterface
{
    /** @var array<string, array<string, array<int, array<int, string|float>>>> */
    private array $values = [];

    /** @var array<string, string> */
    private array $spreadsheetTitles = [];

    /** @var array<string> */
    private array $accessibleSpreadsheets = [];

    /** @var array<string, \Exception> */
    private array $getValuesExceptions = [];

    private string $serviceAccountEmail = 'test@example.com';

    /**
     * @return array<int, array<int, string|float>>|null
     */
    public function getValues(string $spreadsheetId, string $range): ?array
    {
        if (isset($this->getValuesExceptions[$spreadsheetId])) {
            throw $this->getValuesExceptions[$spreadsheetId];
        }

        return $this->values[$spreadsheetId][$range] ?? null;
    }

    public function throwOnGetValues(string $spreadsheetId, \Exception $exception): void
    {
        $this->getValuesExceptions[$spreadsheetId] = $exception;
    }

    /**
     * @param array<int, array<int, string|float>> $values
     */
    public function setValues(string $spreadsheetId, string $range, array $values): void
    {
        if (!isset($this->values[$spreadsheetId])) {
            $this->values[$spreadsheetId] = [];
        }
        $this->values[$spreadsheetId][$range] = $values;
    }

    /**
     * @param array<int, array<int, string|float>> $values
     */
    public function updateValues(string $spreadsheetId, string $range, array $values): void
    {
        $this->setValues($spreadsheetId, $range, $values);
    }

    public function validateSpreadsheetAccess(string $spreadsheetId): bool
    {
        return in_array($spreadsheetId, $this->accessibleSpreadsheets, true);
    }

    public function setSpreadsheetAccessible(string $spreadsheetId, bool $isAccessible = true): void
    {
        if ($isAccessible && !in_array($spreadsheetId, $this->accessibleSpreadsheets, true)) {
            $this->accessibleSpreadsheets[] = $spreadsheetId;
        } elseif (!$isAccessible) {
            $this->accessibleSpreadsheets = array_filter(
                $this->accessibleSpreadsheets,
                fn (string $id) => $id !== $spreadsheetId
            );
        }
    }

    public function getSpreadsheetTitle(string $spreadsheetId): ?string
    {
        return $this->spreadsheetTitles[$spreadsheetId] ?? null;
    }

    public function setSpreadsheetTitle(string $spreadsheetId, string $title): void
    {
        $this->spreadsheetTitles[$spreadsheetId] = $title;
    }

    public function cloneSpreadsheet(string $sourceId, string $newTitle): string
    {
        $newId = $sourceId;
        if (isset($this->values[$sourceId])) {
            $this->values[$newId] = $this->values[$sourceId];
        }
        $this->spreadsheetTitles[$newId] = $newTitle;
        $this->accessibleSpreadsheets[] = $newId;

        return $newId;
    }

    public function getSharingInstructions(string $spreadsheetId): string
    {
        return sprintf(
            'Для работы с таблицей предоставьте доступ на редактирование для %s',
            $this->serviceAccountEmail
        );
    }

    public function addAccessibleSpreadsheet(string $spreadsheetId): void
    {
        $this->accessibleSpreadsheets[] = $spreadsheetId;
        if (!isset($this->spreadsheetTitles[$spreadsheetId])) {
            $this->spreadsheetTitles[$spreadsheetId] = 'Бюджет';
        }
    }

    public function isSpreadsheetAccessible(string $spreadsheetId): bool
    {
        return in_array($spreadsheetId, $this->accessibleSpreadsheets, true);
    }

    public function getServiceAccountEmail(): string
    {
        return $this->serviceAccountEmail;
    }

    public function setServiceAccountEmail(string $email): void
    {
        $this->serviceAccountEmail = $email;
    }
}
