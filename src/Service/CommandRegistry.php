<?php

namespace App\Service;

use App\Service\Command\CommandInterface;

class CommandRegistry
{
    /**
     * @var array<CommandInterface>
     */
    private array $commands;

    /**
     * @param iterable<CommandInterface> $commands
     */
    public function __construct(
        iterable $commands,
    ) {
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
}
