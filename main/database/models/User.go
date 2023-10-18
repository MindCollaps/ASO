package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	Username       string             `json:"username" bson:"username"`
	Password       string             `json:"password" bson:"password"`
	Email          string             `json:"email" bson:"email"`
	GitHubToken    string             `json:"githubToken" bson:"githubToken"`
	GitHubUsername string             `json:"githubUsername" bson:"githubUsername"`
	DateCreated    primitive.DateTime `json:"dateCreated" bson:"dateCreated"`
	IsSuperUser    bool               `json:"isSuperUser" bson:"isSuperUser"`
}
