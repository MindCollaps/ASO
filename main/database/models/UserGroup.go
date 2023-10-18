package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserGroup struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	Date            primitive.DateTime `json:"date" bson:"date"`
	DateExpires     primitive.DateTime `json:"dateExpires" bson:"dateExpires"`
	Expires         bool               `json:"expires" bson:"expires"`
	AutoDelete      bool               `json:"autoDelete" bson:"autoDelete"`
	Notify          bool               `json:"notify" bson:"notify"`
	NotifiedExpired bool               `json:"notified" bson:"notified"`
	GitHubRepo      string             `json:"githubRepo" bson:"githubRepo"`
	Belongs         primitive.ObjectID `json:"belongs" bson:"belongs"`
}
