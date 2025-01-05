<?php

namespace App\Service\Command;

use App\Entity\User;

interface CommandInterface
{
    public function getName(): string;

    public function execute(int $chatId, ?User $user, string $message): void;

    public function supports(string $command): bool;
}
