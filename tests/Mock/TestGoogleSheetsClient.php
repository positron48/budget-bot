<?php

namespace App\Tests\Mock;

use Google\Service\Sheets;
use Google\Service\Sheets\Resource\Spreadsheets;
use Google\Service\Sheets\Resource\SpreadsheetsValues;
use Google\Service\Sheets\ValueRange;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

class TestGoogleSheetsClient extends TestCase
{
    /** @var Sheets&MockObject */
    private Sheets $sheets;
    /** @var Spreadsheets&MockObject */
    private Spreadsheets $spreadsheets;
    /** @var SpreadsheetsValues&MockObject */
    private SpreadsheetsValues $values;

    protected function setUp(): void
    {
        parent::setUp();
        $this->sheets = $this->createMock(Sheets::class);
        $this->spreadsheets = $this->createMock(Spreadsheets::class);
        $this->values = $this->createMock(SpreadsheetsValues::class);

        $this->sheets->spreadsheets = $this->spreadsheets;
        $this->sheets->spreadsheets_values = $this->values;
    }

    public function getSheets(): Sheets
    {
        return $this->sheets;
    }

    /**
     * @param array<int, array<int, string|float>> $values
     */
    public function mockGetValues(string $spreadsheetId, string $range, array $values): void
    {
        $valueRange = new ValueRange();
        $valueRange->setValues($values);

        $this->values->method('get')
            ->with($spreadsheetId, $range)
            ->willReturn($valueRange);
    }

    /**
     * @param array<int, array<int, string|float>> $values
     */
    public function mockUpdateValues(string $spreadsheetId, string $range, array $values): void
    {
        $valueRange = new ValueRange();
        $valueRange->setValues($values);

        $this->values->method('update')
            ->with($spreadsheetId, $range, $valueRange, ['valueInputOption' => 'USER_ENTERED'])
            ->willReturn($valueRange);
    }

    /**
     * @param array<int, array<int, string|float>> $values
     */
    public function mockAppendValues(string $spreadsheetId, string $range, array $values): void
    {
        $valueRange = new ValueRange();
        $valueRange->setValues($values);

        $this->values->method('append')
            ->with($spreadsheetId, $range, $valueRange, ['valueInputOption' => 'USER_ENTERED'])
            ->willReturn($valueRange);
    }
}
