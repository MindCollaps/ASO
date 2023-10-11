package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Token struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Token       string             `json:"token" bson:"token"`
	UserGroup   primitive.ObjectID `json:"userGroup" bson:"userGroup"`
	DateCreated primitive.DateTime `json:"dateCreated" bson:"dateCreated"`
	DateExpires primitive.DateTime `json:"dateExpires" bson:"dateExpires"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	IsReg       bool               `json:"isReg" bson:"isReg"`
}
