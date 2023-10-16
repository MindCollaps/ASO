package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	Username       string             `json:"username" bson:"username"`
	Password       string             `json:"password" bson:"password"`
	GitHubToken    string             `json:"githubToken" bson:"githubToken"`
	GitHubUsername string             `json:"githubUsername" bson:"githubUsername"`
	GitHubRepo     string             `json:"githubRepo" bson:"githubRepo"`
}
