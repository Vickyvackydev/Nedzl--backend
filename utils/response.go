package utils

import "github.com/labstack/echo/v4"

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func ResponseSucess(c echo.Context, status int, message string, data interface{}) error {
	return c.JSON(status, ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	})

}

func ResponseError(c echo.Context, status int, message string, err error) error {
	var errMessage string
	if err != nil {
		errMessage = err.Error()
	} else {
		errMessage = "An error occurred"
	}
	return c.JSON(status, ApiResponse{
		Success: false,
		Message: message,
		Error:   errMessage,
	})
}
