<?php

namespace App\Tests\Entity;

use App\Entity\User;
use PHPUnit\Framework\TestCase;

class UserTest extends TestCase
{
    private User $user;

    protected function setUp(): void
    {
        $this->user = new User();
    }

    public function testGettersAndSetters(): void
    {
        // Test ID (read-only)
        $this->assertNull($this->user->getId());

        // Test telegramId
        $telegramId = 123456789;
        $this->user->setTelegramId($telegramId);
        $this->assertSame($telegramId, $this->user->getTelegramId());

        // Test username
        $username = 'testuser';
        $this->user->setUsername($username);
        $this->assertSame($username, $this->user->getUsername());

        // Test firstName
        $firstName = 'John';
        $this->user->setFirstName($firstName);
        $this->assertSame($firstName, $this->user->getFirstName());

        // Test lastName
        $lastName = 'Doe';
        $this->user->setLastName($lastName);
        $this->assertSame($lastName, $this->user->getLastName());

        // Test currentSpreadsheetId
        $spreadsheetId = 'abc123';
        $this->user->setCurrentSpreadsheetId($spreadsheetId);
        $this->assertSame($spreadsheetId, $this->user->getCurrentSpreadsheetId());

        // Test state
        $state = 'WAITING_SPREADSHEET_ID';
        $this->user->setState($state);
        $this->assertSame($state, $this->user->getState());

        // Test tempData
        $tempData = ['key' => 'value', 'nested' => ['data' => true]];
        $this->user->setTempData($tempData);
        $this->assertSame($tempData, $this->user->getTempData());
    }

    public function testNullableFields(): void
    {
        // Test nullable username
        $this->user->setUsername(null);
        $this->assertNull($this->user->getUsername());

        // Test nullable firstName
        $this->user->setFirstName(null);
        $this->assertNull($this->user->getFirstName());

        // Test nullable lastName
        $this->user->setLastName(null);
        $this->assertNull($this->user->getLastName());

        // Test nullable currentSpreadsheetId
        $this->user->setCurrentSpreadsheetId(null);
        $this->assertNull($this->user->getCurrentSpreadsheetId());
    }

    public function testFluentInterface(): void
    {
        // Test that all setters return $this for method chaining
        $this->assertSame($this->user, $this->user->setTelegramId(123));
        $this->assertSame($this->user, $this->user->setUsername('test'));
        $this->assertSame($this->user, $this->user->setFirstName('John'));
        $this->assertSame($this->user, $this->user->setLastName('Doe'));
        $this->assertSame($this->user, $this->user->setCurrentSpreadsheetId('abc'));
        $this->assertSame($this->user, $this->user->setState('test'));
        $this->assertSame($this->user, $this->user->setTempData([]));
    }

    public function testTempDataDefaultValue(): void
    {
        // Test that getTempData returns empty array by default
        $this->assertSame([], $this->user->getTempData());

        // Test that getTempData still returns empty array when tempData is null
        $this->user->setTempData([]);
        $this->assertSame([], $this->user->getTempData());
    }
}
