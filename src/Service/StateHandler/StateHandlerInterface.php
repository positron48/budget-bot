<?php

namespace App\Service\StateHandler;

use App\Entity\User;

interface StateHandlerInterface
{
    public function supports(string $state): bool;

    /**
     * @return bool Whether the message was handled
     */
    public function handle(int $chatId, User $user, string $message): bool;
}
