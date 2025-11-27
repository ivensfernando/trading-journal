package payloads

import "time"

type AuthPayload struct {
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Bio         string    `json:"bio"`
	AvatarURL   string    `json:"avatar_url"`
	PhoneNumber string    `json:"phone_number"`
	Timezone    time.Time `json:"timezone"`
}
