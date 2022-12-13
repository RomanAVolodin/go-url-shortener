package config

const (
	AppPort string = ":8080"
	BaseUrl string = "http://localhost" + AppPort + "/"
)

const (
	OnlyGetPostRequestAllowedError = "Only GET/POST requests allowed"
	RequestBodyEmptyError          = "Request body is empty"
	UnknownError                   = "Something bad's happened"
	NoIdWasFoundInUrl              = "You should place url id to the url"
	NoUrlFoundById                 = "No url found by id"
)
