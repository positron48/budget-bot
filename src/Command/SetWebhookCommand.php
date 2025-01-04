<?php

namespace App\Command;

use Longman\TelegramBot\Exception\TelegramException;
use Longman\TelegramBot\Telegram;
use Psr\Log\LoggerInterface;
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;

#[AsCommand(
    name: 'app:set-webhook',
    description: 'Sets up the Telegram webhook URL'
)]
class SetWebhookCommand extends Command
{
    private string $botToken;
    private string $botUsername;
    private LoggerInterface $logger;

    public function __construct(
        string $botToken,
        string $botUsername,
        LoggerInterface $logger,
    ) {
        parent::__construct();
        $this->botToken = $botToken;
        $this->botUsername = $botUsername;
        $this->logger = $logger;
    }

    protected function configure(): void
    {
        $this->addArgument('url', InputArgument::REQUIRED, 'The webhook URL');
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        try {
            $telegram = new Telegram($this->botToken, $this->botUsername);
            $result = $telegram->setWebhook($input->getArgument('url'));

            if ($result->isOk()) {
                $this->logger->info('Webhook set successfully', [
                    'url' => $input->getArgument('url'),
                ]);
                $output->writeln('Webhook set successfully!');

                return Command::SUCCESS;
            }

            $this->logger->error('Failed to set webhook', [
                'url' => $input->getArgument('url'),
                'error' => $result->getDescription(),
            ]);
            $output->writeln('Failed to set webhook: '.$result->getDescription());

            return Command::FAILURE;
        } catch (TelegramException $e) {
            $this->logger->error('Error setting webhook: '.$e->getMessage(), [
                'exception' => $e,
                'url' => $input->getArgument('url'),
            ]);
            $output->writeln('Error: '.$e->getMessage());

            return Command::FAILURE;
        }
    }
}
