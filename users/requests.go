package users

type AuthenticationRequest struct {
	EmailAddress string `json:"emailAddress"`
	Password     string `json:"password"`
	Application  string `json:"application,omitempty"`
}

type UpdateRequest struct {
	UserID     string `json:"userid"`
	OptionalID string `json:"optional,omitempty"`
	Field      string `json:"field"`
	Value      string `json:"value"`
}

type AddUserRequest struct {
	EmailAddress string `json:"emailAddress"`
	FirstName    string `json:"firstName"`
	MiddleName   string `json:"middleName,omitempty"`
	LastName     string `json:"lastName"`
	Password     string `json:"password"`
	Application  string `json:"application"`
}

type PasswordResetRequest struct {
	EmailAddress string `json:"emailAddress"`
	Password     string `json:"password"`
	Token        string `json:"token"`
}
