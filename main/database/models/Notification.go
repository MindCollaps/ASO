package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Belongs      primitive.ObjectID `json:"belongs" bson:"belongs"`
	Notification string             `json:"notification" bson:"notification"`
	DateCreated  primitive.DateTime `json:"dateCreated" bson:"dateCreated"`
	Title        string             `json:"title" bson:"title"`
	UserGroup    primitive.ObjectID `json:"userGroup" bson:"userGroup"`
	GitHubUser   primitive.ObjectID `json:"githubUser" bson:"githubUser"`
	Token        primitive.ObjectID `json:"token" bson:"token"`
}
