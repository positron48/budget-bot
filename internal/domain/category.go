// Package domain contains core domain models used across the bot.
package domain

// Category represents an expense/income category with optional emoji.
type Category struct {
    ID    string
    Name  string
    Emoji string
}


