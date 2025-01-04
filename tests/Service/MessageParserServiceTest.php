<?php

namespace App\Tests\Service;

use App\Service\MessageParserService;
use PHPUnit\Framework\TestCase;

class MessageParserServiceTest extends TestCase
{
    private MessageParserService $parser;

    protected function setUp(): void
    {
        $this->parser = new MessageParserService();
    }

    /**
     * @dataProvider validMessageProvider
     *
     * @param array{date: \DateTime, amount: float, description: string, isIncome: bool} $expected
     */
    public function testParseValidMessage(string $message, array $expected): void
    {
        $result = $this->parser->parseMessage($message);

        $this->assertNotNull($result);

        // Сравниваем даты отдельно, так как они могут иметь разное время
        $this->assertInstanceOf(\DateTime::class, $result['date']);
        $this->assertEquals($expected['date']->format('Y-m-d'), $result['date']->format('Y-m-d'));

        // Сравниваем остальные поля
        $this->assertEquals($expected['amount'], $result['amount']);
        $this->assertEquals($expected['description'], $result['description']);
        $this->assertEquals($expected['isIncome'], $result['isIncome']);
    }

    /**
     * @return array<string, array{0: string, 1: array{date: \DateTime, amount: float, description: string, isIncome: bool}}>
     */
    public function validMessageProvider(): array
    {
        return [
            'simple expense' => [
                '100 продукты',
                [
                    'date' => new \DateTime('today'),
                    'amount' => 100.0,
                    'description' => 'продукты',
                    'isIncome' => false,
                ],
            ],
            'expense with date' => [
                '01.01.2024 200 такси',
                [
                    'date' => new \DateTime('2024-01-01'),
                    'amount' => 200.0,
                    'description' => 'такси',
                    'isIncome' => false,
                ],
            ],
            'income with plus' => [
                '+5000 зарплата',
                [
                    'date' => new \DateTime('today'),
                    'amount' => 5000.0,
                    'description' => 'зарплата',
                    'isIncome' => true,
                ],
            ],
            'income with date' => [
                '31.12.2023 +10000 премия',
                [
                    'date' => new \DateTime('2023-12-31'),
                    'amount' => 10000.0,
                    'description' => 'премия',
                    'isIncome' => true,
                ],
            ],
            'decimal amount with dot' => [
                '99.90 кофе',
                [
                    'date' => new \DateTime('today'),
                    'amount' => 99.90,
                    'description' => 'кофе',
                    'isIncome' => false,
                ],
            ],
            'decimal amount with comma' => [
                '99,90 кофе',
                [
                    'date' => new \DateTime('today'),
                    'amount' => 99.90,
                    'description' => 'кофе',
                    'isIncome' => false,
                ],
            ],
        ];
    }

    /**
     * @dataProvider invalidMessageProvider
     */
    public function testParseInvalidMessage(string $message): void
    {
        $result = $this->parser->parseMessage($message);
        $this->assertNull($result);
    }

    /**
     * @return array<string, array{0: string}>
     */
    public function invalidMessageProvider(): array
    {
        return [
            'empty message' => [''],
            'only amount' => ['100'],
            'only text' => ['продукты'],
            'invalid date format' => ['24.01.01 100 продукты'],
            'invalid amount' => ['abc 100 продукты'],
            'missing description' => ['100'],
            'missing amount' => ['продукты'],
            'invalid year format' => ['01.01.24 100 продукты'],
            'missing description after amount' => ['100 '],
            'missing description after decimal' => ['99.90 '],
        ];
    }
}
