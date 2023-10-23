package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Belongs      primitive.ObjectID `json:"belongs" bson:"belongs"`
	Notification string             `json:"notification" bson:"notification"`
	DateCreated  primitive.DateTime `json:"dateCreated" bson:"dateCreated"`
	Title        string             `json:"title" bson:"title"`
	Style        string             `json:"style" bson:"style"`
	UserGroup    primitive.ObjectID `json:"userGroup" bson:"userGroup"`
	GitHubUser   primitive.ObjectID `json:"githubUser" bson:"githubUser"`
	Profile      bool               `json:"profile" bson:"profile"`
	Token        primitive.ObjectID `json:"token" bson:"token"`
	Seen         bool               `json:"seen" bson:"seen"`
	Popup        bool               `json:"popup" bson:"popup"`
}
