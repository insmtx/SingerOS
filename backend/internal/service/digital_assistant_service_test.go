package service

import (
	"testing"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDigitalAssistantDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&types.DigitalAssistant{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func TestCreateDigitalAssistant_ValidInput(t *testing.T) {
	db := setupDigitalAssistantDB(t)
	ctx := setupTestContextWithCaller(t)

	service := NewDigitalAssistantService(db, nil)

	req := &contract.CreateDigitalAssistantRequest{
		Code:         "test-code",
		Name:         "Test Name",
		Description:  "Test Description",
		SystemPrompt: "You are a test assistant",
	}

	result, err := service.CreateDigitalAssistant(ctx, req)
	if err != nil {
		t.Fatalf("CreateDigitalAssistant failed: %v", err)
	}

	if result.Code != "test-code" {
		t.Errorf("expected code test-code, got %s", result.Code)
	}
	if result.Name != "Test Name" {
		t.Errorf("expected name 'Test Name', got %s", result.Name)
	}
}

func TestCreateDigitalAssistant_WithoutCaller(t *testing.T) {
	db := setupDigitalAssistantDB(t)
	ctx := setupTestContextWithoutCaller(t)

	service := NewDigitalAssistantService(db, nil)

	req := &contract.CreateDigitalAssistantRequest{
		Code: "test-code",
		Name: "Test Name",
	}

	_, err := service.CreateDigitalAssistant(ctx, req)
	if err == nil {
		t.Fatal("expected error when caller is not in context")
	}
	if err.Error() != "user not authenticated or org not set" {
		t.Errorf("expected 'user not authenticated or org not set', got %v", err)
	}
}

func TestCreateDigitalAssistant_MissingCode(t *testing.T) {
	db := setupDigitalAssistantDB(t)
	ctx := setupTestContextWithCaller(t)

	service := NewDigitalAssistantService(db, nil)

	req := &contract.CreateDigitalAssistantRequest{
		Name: "Test Name",
	}

	_, err := service.CreateDigitalAssistant(ctx, req)
	if err == nil {
		t.Fatal("expected error when code is missing")
	}
}

func TestCreateDigitalAssistant_MissingName(t *testing.T) {
	db := setupDigitalAssistantDB(t)
	ctx := setupTestContextWithCaller(t)

	service := NewDigitalAssistantService(db, nil)

	req := &contract.CreateDigitalAssistantRequest{
		Code: "test-code",
	}

	_, err := service.CreateDigitalAssistant(ctx, req)
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
}
