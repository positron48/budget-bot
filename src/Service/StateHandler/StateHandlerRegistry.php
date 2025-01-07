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
            $this->logger->debug('Checking handler', [
                'handler_class' => get_class($handler),
                'supports_state' => $handler->supports($state),
            ]);
            if ($handler->supports($state)) {
                $result = $handler->handle($chatId, $user, $message);
                $this->logger->debug('Handler result', [
                    'handler_class' => get_class($handler),
                    'result' => $result,
                ]);

                return $result;
            }
        }

        return false;
    }
}
