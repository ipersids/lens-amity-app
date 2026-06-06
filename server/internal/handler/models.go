package handler

import (
	"lensamity/internal/db"
)

type Env struct {
	*db.Store
}

// DTO -->

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SignupRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type SignupReponse struct {
	Data SignupReponseBody `json:"data"`
}

type SignupReponseBody struct {
	Uuid        string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type UserProfileReponse struct {
	Data UserProfileReponseBody `json:"data"`
}

type UserProfileReponseBody struct {
	Uuid        string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}
