<?php

namespace App\Tests\Service;

use App\Service\MessageParser;
use Longman\TelegramBot\Entities\Message;
use PHPUnit\Framework\TestCase;

class MessageParserTest extends TestCase
{
    private MessageParser $parser;

    protected function setUp(): void
    {
        $this->parser = new MessageParser();
    }

    public function testParseEmptyMessage(): void
    {
        $message = $this->createMock(Message::class);
        $message->expects($this->once())
            ->method('getText')
            ->willReturn('');

        $result = $this->parser->parse($message);

        $this->assertEquals([
            'command' => '',
            'arguments' => [],
        ], $result);
    }

    public function testParseCommandWithoutArguments(): void
    {
        $message = $this->createMock(Message::class);
        $message->expects($this->once())
            ->method('getText')
            ->willReturn('/start');

        $result = $this->parser->parse($message);

        $this->assertEquals([
            'command' => '/start',
            'arguments' => [],
        ], $result);
    }

    public function testParseCommandWithArguments(): void
    {
        $message = $this->createMock(Message::class);
        $message->expects($this->once())
            ->method('getText')
            ->willReturn('/command arg1 arg2 arg3');

        $result = $this->parser->parse($message);

        $this->assertEquals([
            'command' => '/command',
            'arguments' => ['arg1', 'arg2', 'arg3'],
        ], $result);
    }

    public function testParseWithExtraSpaces(): void
    {
        $message = $this->createMock(Message::class);
        $message->expects($this->once())
            ->method('getText')
            ->willReturn('  /command   arg1    arg2   ');

        $result = $this->parser->parse($message);

        $this->assertEquals([
            'command' => '/command',
            'arguments' => ['arg1', 'arg2'],
        ], $result);
    }
}
