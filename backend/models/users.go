package models 

type CreateUserRequest struct {
	UserID    string  `json:"user_id" binding: "required"`
	Name	  string  `json: "name"   binding: "required"`
}

type UserResponse struct {
	UserID     string 	`json: "user_id"`
	Name       string   `json: "name"`
}

type FollowRequest struct {
	FollowerID  string   `json: "follower_id" binding: "required"`
	FolloweeID  string   `jsno: "followee_id" binding: "required"`
}