package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserGroup struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	Date            primitive.DateTime `json:"date" bson:"date"`
	DateExpires     primitive.DateTime `json:"dateExpires" bson:"dateExpires"`
	Expires         bool               `json:"expires" bson:"expires"`
	AutoRemoveUsers bool               `json:"autoRemoveUsers" bson:"autoRemoveUsers"`
	AutoDelete      bool               `json:"autoDelete" bson:"autoDelete"`
	Notify          bool               `json:"notify" bson:"notify"`
	NotifiedExpired bool               `json:"notifiedExpired" bson:"notifiedExpired"`
	NotifiedDeleted bool               `json:"notifiedDeleted" bson:"notifiedDeleted"`
	GitHubRepo      string             `json:"githubRepo" bson:"githubRepo"`
	GitHubOwner     string             `json:"githubOwner" bson:"githubOwner"`
	Belongs         primitive.ObjectID `json:"belongs" bson:"belongs"`
}
