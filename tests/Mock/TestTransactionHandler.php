<?php

namespace App\Tests\Mock;

use App\Repository\UserRepository;
use App\Service\CategoryService;
use App\Service\GoogleSheetsService;
use App\Service\TelegramApiServiceInterface;
use App\Service\TransactionHandler;
use Psr\Log\LoggerInterface;

class TestTransactionHandler extends TransactionHandler
{
    public function __construct(
        GoogleSheetsService $sheetsService,
        CategoryService $categoryService,
        LoggerInterface $logger,
        UserRepository $userRepository,
        TelegramApiServiceInterface $telegramApi,
    ) {
        parent::__construct($sheetsService, $categoryService, $logger, $userRepository, $telegramApi);
    }
}
