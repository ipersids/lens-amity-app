package handler

import (
	"lensamity/internal/core"
	"lensamity/internal/db"
)

type Env struct {
	Store  *db.Store
	Config *core.Config
}

// DTO -->

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}

type SignupRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponseBody struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         UserProfileReponseBody
}

type UserProfileReponseBody struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}
