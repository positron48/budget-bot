<?php

namespace App\Service;

use App\Entity\User;
use App\Service\Command\CommandInterface;

class CommandRegistry
{
    /** @var CommandInterface[] */
    private array $commands;

    /**
     * @param CommandInterface[] $commands
     */
    public function __construct(iterable $commands)
    {
        $this->commands = [];
        foreach ($commands as $command) {
            $this->commands[] = $command;
        }
    }

    public function findCommand(string $message): ?CommandInterface
    {
        foreach ($this->commands as $command) {
            if ($command->supports($message)) {
                return $command;
            }
        }

        return null;
    }

    public function executeCommand(CommandInterface $command, int $chatId, ?User $user, string $message): void
    {
        $command->execute($chatId, $user, $message);
    }
}
