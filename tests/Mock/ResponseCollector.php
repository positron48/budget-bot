<?php

namespace App\Tests\Mock;

class ResponseCollector
{
    private static ?self $instance = null;
    /** @var array<int, string> */
    private array $responses = [];

    private function __construct()
    {
    }

    public static function getInstance(): self
    {
        if (null === self::$instance) {
            self::$instance = new self();
        }

        return self::$instance;
    }

    public function addResponse(string $text): void
    {
        $this->responses[] = $text;
    }

    /** @return array<int, string> */
    public function getResponses(): array
    {
        return $this->responses;
    }

    public function reset(): void
    {
        $this->responses = [];
    }

    public static function resetInstance(): void
    {
        self::$instance = null;
    }
}
