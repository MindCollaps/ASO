package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type GitHubUser struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	GitHubUsername string             `json:"githubUsername" bson:"githubUsername"`
	GitHubID       string             `json:"githubID" bson:"githubID"`
	DateCreated    primitive.DateTime `json:"dateCreated" bson:"dateCreated"`
	DateExpires    primitive.DateTime `json:"dateExpires" bson:"dateExpires"`
	UserGroup      primitive.ObjectID `json:"userGroup" bson:"userGroup"`
}
