<?php

namespace App\Service\StateHandler;

use App\Entity\User;
use Psr\Log\LoggerInterface;

class StateHandlerRegistry
{
    /**
     * @var array<StateHandlerInterface>
     */
    private array $handlers;

    private LoggerInterface $logger;

    /**
     * @param iterable<StateHandlerInterface> $handlers
     */
    public function __construct(iterable $handlers, LoggerInterface $logger)
    {
        $this->handlers = [];
        foreach ($handlers as $handler) {
            $this->handlers[] = $handler;
        }
        $this->logger = $logger;
    }

    public function handleState(int $chatId, User $user, string $message): bool
    {
        $state = $user->getState();
        if (!$state) {
            return false;
        }

        $this->logger->info('Handling state', [
            'chat_id' => $chatId,
            'state' => $state,
            'message' => $message,
        ]);

        foreach ($this->handlers as $handler) {
            if ($handler->supports($state)) {
                return $handler->handle($chatId, $user, $message);
            }
        }

        return false;
    }
}
