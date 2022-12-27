package config

const (
	AppPort string = ":8080"
	BaseURL        = "http://localhost" + AppPort + "/"
)

const (
	OnlyGetPostRequestAllowedError = "Only GET/POST requests allowed"
	RequestBodyEmptyError          = "Request body is empty"
	BadInputData                   = "Incorrect request body data"
	UnknownError                   = "Something bad's happened"
	NoURLFoundByID                 = "No url found by id"
)
