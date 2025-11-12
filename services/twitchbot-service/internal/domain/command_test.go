package domain

import (
	"testing"
)

func TestCommandTableName(t *testing.T) {
	command := Command{}
	expected := "commands"
	if got := command.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestCommandCreation(t *testing.T) {
	command := Command{
		ID:              "test-id",
		Name:            "!hello",
		Description:     "Says hello",
		Response:        "Hello, @{username}!",
		IsActive:        true,
		ModeratorOnly:   false,
		SubscriberOnly:  false,
		CooldownSeconds: 30,
		UsageCount:      0,
	}

	if command.Name != "!hello" {
		t.Errorf("Name = %v, want %v", command.Name, "!hello")
	}

	if command.Response != "Hello, @{username}!" {
		t.Errorf("Response = %v, want %v", command.Response, "Hello, @{username}!")
	}

	if !command.IsActive {
		t.Error("IsActive should be true")
	}

	if command.ModeratorOnly {
		t.Error("ModeratorOnly should be false")
	}

	if command.CooldownSeconds != 30 {
		t.Errorf("CooldownSeconds = %v, want %v", command.CooldownSeconds, 30)
	}
}
