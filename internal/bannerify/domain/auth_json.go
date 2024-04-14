package domain

type AuthorizationData struct {
	Login    string `example:"login"    json:"login"`
	Password string `example:"password" json:"password"`
}

type Token struct {
	Token string `json:"token"`
}
