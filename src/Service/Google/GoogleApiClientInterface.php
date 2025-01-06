<?php

namespace App\Service\Google;

interface GoogleApiClientInterface
{
    /**
     * @return array<int, array<int, string|float>>|null
     */
    public function getValues(string $spreadsheetId, string $range): ?array;

    /**
     * @param array<int, array<int, string|float>> $values
     */
    public function updateValues(string $spreadsheetId, string $range, array $values): void;

    public function validateSpreadsheetAccess(string $spreadsheetId): bool;

    public function getSpreadsheetTitle(string $spreadsheetId): ?string;

    public function cloneSpreadsheet(string $sourceId, string $newTitle): string;

    public function getSharingInstructions(string $spreadsheetId): string;
}
