package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type GitHubUser struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	Username       string             `json:"username" bson:"username"`
	GitHubUsername string             `json:"githubUsername" bson:"githubUsername"`
	DateCreated    primitive.DateTime `json:"dateCreated" bson:"dateCreated"`
	Expires        bool               `json:"expires" bson:"expires"`
	ExpiresGroup   bool               `json:"expiresGroup" bson:"expiresGroup"`
	DateExpires    primitive.DateTime `json:"dateExpires" bson:"dateExpires"`
	UserGroup      primitive.ObjectID `json:"userGroup" bson:"userGroup"`
	AddedToRepo    bool               `json:"addedToRepo" bson:"addedToRepo"`
	Belongs        primitive.ObjectID `json:"belongs" bson:"belongs"`
}
