<?php

namespace App\Service\StateHandler;

use App\Entity\User;

class StateHandlerRegistry
{
    /** @var StateHandlerInterface[] */
    private array $handlers;

    /**
     * @param StateHandlerInterface[] $handlers
     */
    public function __construct(iterable $handlers)
    {
        $this->handlers = [];
        foreach ($handlers as $handler) {
            $this->handlers[] = $handler;
        }
    }

    public function findHandler(string $state): ?StateHandlerInterface
    {
        foreach ($this->handlers as $handler) {
            if ($handler->supports($state)) {
                return $handler;
            }
        }

        return null;
    }

    public function handleState(int $chatId, User $user, string $message): void
    {
        $state = $user->getState();
        if (!$state) {
            return;
        }

        $handler = $this->findHandler($state);
        if ($handler) {
            $handler->handle($chatId, $user, $message);
        }
    }
} 