<?php

namespace App\Service;

use App\Service\Command\CategoriesCommand;
use App\Service\Command\ClearCategoriesCommand;
use App\Service\Command\CommandInterface;
use App\Service\Command\ListCommand;
use App\Service\Command\MapCommand;
use App\Service\Command\RemoveCommand;
use App\Service\Command\StartCommand;
use App\Service\Command\SyncCategoriesCommand;

class CommandRegistry
{
    /**
     * @var array<CommandInterface>
     */
    private array $commands;

    public function __construct(
        StartCommand $startCommand,
        ListCommand $listCommand,
        RemoveCommand $removeCommand,
        CategoriesCommand $categoriesCommand,
        MapCommand $mapCommand,
        SyncCategoriesCommand $syncCategoriesCommand,
        ClearCategoriesCommand $clearCategoriesCommand,
    ) {
        $this->commands = [
            $startCommand,
            $listCommand,
            $removeCommand,
            $categoriesCommand,
            $mapCommand,
            $syncCategoriesCommand,
            $clearCategoriesCommand,
        ];
    }

    public function findCommand(string $message): ?CommandInterface
    {
        foreach ($this->commands as $command) {
            if (str_starts_with($message, $command->getName())) {
                return $command;
            }
        }

        return null;
    }
}
