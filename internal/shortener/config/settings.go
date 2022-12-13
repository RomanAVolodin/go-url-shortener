package config

const (
	AppPort string = ":8080"
	BaseURL string = "http://localhost" + AppPort + "/"
)

const (
	OnlyGetPostRequestAllowedError = "Only GET/POST requests allowed"
	RequestBodyEmptyError          = "Request body is empty"
	UnknownError                   = "Something bad's happened"
	NoIDWasFoundInURL              = "You should place url id to the url"
	NoURLFoundByID                 = "No url found by id"
)
