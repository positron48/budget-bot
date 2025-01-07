<?php

namespace App\Tests\Service;

use App\Entity\User;
use App\Service\StateHandler\StateHandlerInterface;
use App\Service\StateHandler\StateHandlerRegistry;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class StateHandlerRegistryTest extends TestCase
{
    private const TEST_CHAT_ID = 123456;

    private StateHandlerRegistry $registry;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var StateHandlerInterface&MockObject */
    private StateHandlerInterface $handler1;
    /** @var StateHandlerInterface&MockObject */
    private StateHandlerInterface $handler2;

    protected function setUp(): void
    {
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->handler1 = $this->createMock(StateHandlerInterface::class);
        $this->handler2 = $this->createMock(StateHandlerInterface::class);

        $this->registry = new StateHandlerRegistry([$this->handler1, $this->handler2], $this->logger);
    }

    public function testHandleStateWithNoState(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('');

        $this->handler1->expects($this->never())
            ->method('supports');
        $this->handler2->expects($this->never())
            ->method('supports');

        $result = $this->registry->handleState(self::TEST_CHAT_ID, $user, 'test message');
        $this->assertFalse($result);
    }

    public function testHandleStateWithNoSupportingHandler(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('TEST_STATE');

        $this->handler1->expects($this->atLeastOnce())
            ->method('supports')
            ->with('TEST_STATE')
            ->willReturn(false);

        $this->handler2->expects($this->atLeastOnce())
            ->method('supports')
            ->with('TEST_STATE')
            ->willReturn(false);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Handling state', [
                'chat_id' => self::TEST_CHAT_ID,
                'state' => 'TEST_STATE',
                'message' => 'test message',
            ]);

        $result = $this->registry->handleState(self::TEST_CHAT_ID, $user, 'test message');
        $this->assertFalse($result);
    }

    public function testHandleStateWithSupportingHandler(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('TEST_STATE');

        $this->handler1->expects($this->atLeastOnce())
            ->method('supports')
            ->with('TEST_STATE')
            ->willReturn(true);

        $this->handler1->expects($this->once())
            ->method('handle')
            ->with(self::TEST_CHAT_ID, $user, 'test message')
            ->willReturn(true);

        $this->handler2->expects($this->never())
            ->method('supports');

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Handling state', [
                'chat_id' => self::TEST_CHAT_ID,
                'state' => 'TEST_STATE',
                'message' => 'test message',
            ]);

        $result = $this->registry->handleState(self::TEST_CHAT_ID, $user, 'test message');
        $this->assertTrue($result);
    }

    public function testHandleStateWithSecondSupportingHandler(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('TEST_STATE');

        $this->handler1->expects($this->atLeastOnce())
            ->method('supports')
            ->with('TEST_STATE')
            ->willReturn(false);

        $this->handler2->expects($this->atLeastOnce())
            ->method('supports')
            ->with('TEST_STATE')
            ->willReturn(true);

        $this->handler2->expects($this->once())
            ->method('handle')
            ->with(self::TEST_CHAT_ID, $user, 'test message')
            ->willReturn(true);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Handling state', [
                'chat_id' => self::TEST_CHAT_ID,
                'state' => 'TEST_STATE',
                'message' => 'test message',
            ]);

        $result = $this->registry->handleState(self::TEST_CHAT_ID, $user, 'test message');
        $this->assertTrue($result);
    }

    public function testHandleStateWithHandlerReturningFalse(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('TEST_STATE');

        $this->handler1->expects($this->atLeastOnce())
            ->method('supports')
            ->with('TEST_STATE')
            ->willReturn(true);

        $this->handler1->expects($this->once())
            ->method('handle')
            ->with(self::TEST_CHAT_ID, $user, 'test message')
            ->willReturn(false);

        $this->handler2->expects($this->never())
            ->method('supports');

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Handling state', [
                'chat_id' => self::TEST_CHAT_ID,
                'state' => 'TEST_STATE',
                'message' => 'test message',
            ]);

        $result = $this->registry->handleState(self::TEST_CHAT_ID, $user, 'test message');
        $this->assertFalse($result);
    }
}
