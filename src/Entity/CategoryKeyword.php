<?php

namespace App\Entity;

use App\Repository\CategoryKeywordRepository;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity(repositoryClass: CategoryKeywordRepository::class)]
class CategoryKeyword
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    private ?int $id = null;

    #[ORM\ManyToOne(targetEntity: Category::class, inversedBy: 'keywords')]
    private ?Category $category = null;

    #[ORM\ManyToOne(targetEntity: UserCategory::class, inversedBy: 'keywords')]
    private ?UserCategory $userCategory = null;

    #[ORM\Column(length: 255)]
    private ?string $keyword = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getCategory(): ?Category
    {
        return $this->category;
    }

    public function setCategory(?Category $category): static
    {
        $this->category = $category;

        return $this;
    }

    public function getUserCategory(): ?UserCategory
    {
        return $this->userCategory;
    }

    public function setUserCategory(?UserCategory $userCategory): static
    {
        $this->userCategory = $userCategory;

        return $this;
    }

    public function getKeyword(): ?string
    {
        return $this->keyword;
    }

    public function setKeyword(string $keyword): static
    {
        $this->keyword = $keyword;

        return $this;
    }
}
