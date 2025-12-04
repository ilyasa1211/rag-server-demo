package dto

type ErrorResponse struct {
	Code   int    `json:"code"`
	Detail string `json:"detail"`
}
