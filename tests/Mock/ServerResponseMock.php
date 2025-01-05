<?php

namespace App\Tests\Mock;

use Longman\TelegramBot\Entities\ServerResponse;

class ServerResponseMock extends ServerResponse
{
    protected bool $ok = true;
    /** @var array<string, mixed> */
    protected array $raw_data = [];
    protected string $bot_username = '';
    protected ?string $result = null;
    protected ?string $description = null;
    protected ?int $error_code = null;

    /**
     * @param array<string, mixed> $data
     */
    public function __construct(array $data = ['ok' => true])
    {
        $this->raw_data = $data;
        $this->ok = $data['ok'] ?? true;
        $this->result = $data['result'] ?? null;
        $this->description = $data['description'] ?? null;
        $this->error_code = $data['error_code'] ?? null;
    }

    public function isOk(): bool
    {
        return $this->ok;
    }

    public function getResult(): mixed
    {
        return $this->result;
    }

    public function getDescription(): ?string
    {
        return $this->description;
    }

    public function getErrorCode(): ?int
    {
        return $this->error_code;
    }
}
