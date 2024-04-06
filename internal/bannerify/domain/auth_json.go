package domain

type AuthorizationData struct {
	Login    string `json:"login" example:"login"`
	Password string `json:"password" example:"password"`
}

type Token struct {
	Token string `json:"token"`
}
