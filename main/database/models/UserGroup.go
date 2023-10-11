package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserGroup struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Date        primitive.DateTime `json:"date" bson:"date"`
	DateExpires primitive.DateTime `json:"dateExpires" bson:"dateExpires"`
}
