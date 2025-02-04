<?php

namespace App\Utility;

class DateTimeUtility
{
    protected ?\DateTime $mockedDate = null;
    protected ?\DateTime $currentDate = null;

    public static function createWithFixedDate(\DateTime $date): self
    {
        $instance = new self();
        $instance->setCurrentDate(clone $date);

        return $instance;
    }

    public function getCurrentDate(): \DateTime
    {
        if (null !== $this->currentDate) {
            return clone $this->currentDate;
        }

        return new \DateTime();
    }

    public function setCurrentDate(\DateTime $date): void
    {
        $this->currentDate = clone $date;
    }

    public function resetCurrentDate(): void
    {
        $this->currentDate = null;
    }
}
