<?php

namespace App\Tests\Service;

use App\Service\MessageParserService;
use PHPUnit\Framework\TestCase;

/**
 * @covers \App\Service\MessageParserService
 */
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
        $this->assertEquals($expected['date']->format('Y-m-d'), $result['date']->format('Y-m-d'));
        $this->assertEquals($expected['amount'], $result['amount']);
        $this->assertEquals($expected['description'], $result['description']);
        $this->assertEquals($expected['isIncome'], $result['isIncome']);
    }

    /**
     * @return array<string, array{0: string, 1: array{date: \DateTime, amount: float, description: string, isIncome: bool}}>
     */
    public function validMessageProvider(): array
    {
        $today = new \DateTime();
        $yesterday = new \DateTime('-1 day');

        return [
            'expense without date' => [
                '200 такси',
                [
                    'date' => $today,
                    'amount' => 200.0,
                    'description' => 'такси',
                    'isIncome' => false,
                ],
            ],
            'income without date' => [
                '+10000 премия',
                [
                    'date' => $today,
                    'amount' => 10000.0,
                    'description' => 'премия',
                    'isIncome' => true,
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
            'income with date' => [
                '31.12.2023 +10000 премия',
                [
                    'date' => new \DateTime('2023-12-31'),
                    'amount' => 10000.0,
                    'description' => 'премия',
                    'isIncome' => true,
                ],
            ],
            'expense with comma' => [
                '99,90 кофе',
                [
                    'date' => $today,
                    'amount' => 99.90,
                    'description' => 'кофе',
                    'isIncome' => false,
                ],
            ],
            'expense with dot' => [
                '99.90 кофе',
                [
                    'date' => $today,
                    'amount' => 99.90,
                    'description' => 'кофе',
                    'isIncome' => false,
                ],
            ],
            'expense with today keyword' => [
                'сегодня 150 обед',
                [
                    'date' => $today,
                    'amount' => 150.0,
                    'description' => 'обед',
                    'isIncome' => false,
                ],
            ],
            'expense with yesterday keyword' => [
                'вчера 300 ужин',
                [
                    'date' => $yesterday,
                    'amount' => 300.0,
                    'description' => 'ужин',
                    'isIncome' => false,
                ],
            ],
            'expense with d/m/Y format' => [
                '01/01/2024 200 такси',
                [
                    'date' => new \DateTime('2024-01-01'),
                    'amount' => 200.0,
                    'description' => 'такси',
                    'isIncome' => false,
                ],
            ],
            'expense with d/m format' => [
                '01/01 200 такси',
                [
                    'date' => (new \DateTime())->setDate((int) $today->format('Y'), 1, 1),
                    'amount' => 200.0,
                    'description' => 'такси',
                    'isIncome' => false,
                ],
            ],
            'expense with d.m format' => [
                '01.01 200 такси',
                [
                    'date' => (new \DateTime())->setDate((int) $today->format('Y'), 1, 1),
                    'amount' => 200.0,
                    'description' => 'такси',
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
        $this->assertNull($this->parser->parseMessage($message));
    }

    /**
     * @return array<string, array{0: string}>
     */
    public function invalidMessageProvider(): array
    {
        return [
            'empty string' => [''],
            'only date' => ['01.01.2024'],
            'only amount' => ['100'],
            'only description' => ['такси'],
            'invalid date' => ['32.13.2024 100 такси'],
            'invalid amount' => ['01.01.2024 abc такси'],
            'zero amount' => ['01.01.2024 0 такси'],
            'negative amount' => ['01.01.2024 -100 такси'],
            'invalid year format' => ['01.01.24 100 такси'],
            'invalid year length' => ['01.01.20244 100 такси'],
            'invalid month' => ['01.13.2024 100 такси'],
            'invalid day for month' => ['31.04.2024 100 такси'],
            'invalid date format' => ['2024.01.01 100 такси'],
            'invalid date separator' => ['01-01-2024 100 такси'],
        ];
    }

    public function testParseMessageWithLargeNumber(): void
    {
        $result = $this->parser->parseMessage('1000 продукты');

        $this->assertNotNull($result);
        $this->assertInstanceOf(\DateTime::class, $result['date']);
        $this->assertEquals(1000.0, $result['amount']);
        $this->assertEquals('продукты', $result['description']);
        $this->assertFalse($result['isIncome']);
    }

    public function testParseMessageWithDateLikeNumber(): void
    {
        $result = $this->parser->parseMessage('1.00 продукты');

        $this->assertNotNull($result);
        $this->assertInstanceOf(\DateTime::class, $result['date']);
        $this->assertEquals(1.0, $result['amount']);
        $this->assertEquals('продукты', $result['description']);
        $this->assertFalse($result['isIncome']);
    }

    /**
     * @dataProvider validDateProvider
     */
    public function testParseDate(string $dateStr, \DateTime $expected): void
    {
        $result = $this->parser->parseDate($dateStr);
        $this->assertNotNull($result);
        $this->assertEquals($expected->format('Y-m-d'), $result->format('Y-m-d'));
    }

    /**
     * @return array<string, array{0: string, 1: \DateTime}>
     */
    public function validDateProvider(): array
    {
        $today = new \DateTime();
        $yesterday = new \DateTime('-1 day');

        return [
            'today keyword' => ['сегодня', $today],
            'yesterday keyword' => ['вчера', $yesterday],
            'd.m.Y format' => ['01.01.2024', new \DateTime('2024-01-01')],
            'd.m format' => ['01.01', (new \DateTime())->setDate((int) $today->format('Y'), 1, 1)],
            'd/m/Y format' => ['01/01/2024', new \DateTime('2024-01-01')],
            'd/m format' => ['01/01', (new \DateTime())->setDate((int) $today->format('Y'), 1, 1)],
        ];
    }

    /**
     * @dataProvider invalidDateProvider
     */
    public function testParseDateWithInvalidInput(string $dateStr): void
    {
        $result = $this->parser->parseDate($dateStr);
        $this->assertNull($result);
    }

    /**
     * @return array<string, array{0: string}>
     */
    public function invalidDateProvider(): array
    {
        return [
            'empty string' => [''],
            'invalid format' => ['2024.01.01'],
            'invalid separator' => ['01-01-2024'],
            'invalid day' => ['32.01.2024'],
            'invalid month' => ['01.13.2024'],
            'invalid year format' => ['01.01.24'],
            'invalid year length' => ['01.01.20244'],
            'invalid day for month' => ['31.04.2024'],
            'random text' => ['some text'],
        ];
    }
}
