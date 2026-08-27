package xfiber

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestDefaultErrorHandler_FiberError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: DefaultErrorHandler(),
	})
	app.Get("/test", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "bad request msg")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestDefaultErrorHandler_NonFiberError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: DefaultErrorHandler(),
	})
	app.Get("/test", func(c fiber.Ctx) error {
		return errors.New("some internal error")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["message"] != "some internal error" {
		t.Errorf("expected message 'some internal error', got %v", body["message"])
	}
}

func TestServerError(t *testing.T) {
	t.Run("no message", func(t *testing.T) {
		err := ServerError()
		if err.Code != fiber.StatusInternalServerError {
			t.Errorf("expected code 500, got %d", err.Code)
		}
	})

	t.Run("with message", func(t *testing.T) {
		err := ServerError("custom error")
		if err.Code != fiber.StatusInternalServerError {
			t.Errorf("expected code 500, got %d", err.Code)
		}
		if err.Message != "custom error" {
			t.Errorf("expected 'custom error', got '%s'", err.Message)
		}
	})
}

func TestAuthorizedError(t *testing.T) {
	t.Run("no message", func(t *testing.T) {
		err := AuthorizedError()
		if err.Code != fiber.StatusUnauthorized {
			t.Errorf("expected code 401, got %d", err.Code)
		}
	})

	t.Run("with message", func(t *testing.T) {
		err := AuthorizedError("token expired")
		if err.Code != fiber.StatusUnauthorized {
			t.Errorf("expected code 401, got %d", err.Code)
		}
		if err.Message != "token expired" {
			t.Errorf("expected 'token expired', got '%s'", err.Message)
		}
	})
}

func TestParamError(t *testing.T) {
	t.Run("no message", func(t *testing.T) {
		err := ParamError()
		if err.Code != fiber.StatusUnprocessableEntity {
			t.Errorf("expected code 422, got %d", err.Code)
		}
		if err.Message != "Param Error" {
			t.Errorf("expected 'Param Error', got '%s'", err.Message)
		}
	})

	t.Run("with message", func(t *testing.T) {
		err := ParamError("field required")
		if err.Code != fiber.StatusUnprocessableEntity {
			t.Errorf("expected code 422, got %d", err.Code)
		}
		if err.Message != "field required" {
			t.Errorf("expected 'field required', got '%s'", err.Message)
		}
	})
}

func TestTipError(t *testing.T) {
	t.Run("no message", func(t *testing.T) {
		err := TipError()
		if err.Code != fiber.StatusBadRequest {
			t.Errorf("expected code 400, got %d", err.Code)
		}
		if err.Message != "Tip Error" {
			t.Errorf("expected 'Tip Error', got '%s'", err.Message)
		}
	})

	t.Run("with message", func(t *testing.T) {
		err := TipError("参数不能为空")
		if err.Code != fiber.StatusBadRequest {
			t.Errorf("expected code 400, got %d", err.Code)
		}
		if err.Message != "参数不能为空" {
			t.Errorf("expected '参数不能为空', got '%s'", err.Message)
		}
	})
}
