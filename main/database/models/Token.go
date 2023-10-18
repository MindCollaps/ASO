package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Token struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id"`
	Name                    string             `json:"name" bson:"name"`
	Count                   int                `json:"count" bson:"count"`
	Token                   string             `json:"token" bson:"token"`
	UserGroup               primitive.ObjectID `json:"userGroup" bson:"userGroup"`
	DateCreated             primitive.DateTime `json:"dateCreated" bson:"dateCreated"`
	DateExpires             primitive.DateTime `json:"dateExpires" bson:"dateExpires"`
	CreatedBy               primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	IsUserRegistrationToken bool               `json:"isReg" bson:"isReg"`
	SuperUser               bool               `json:"superUser" bson:"superUser"`
	DirectAdd               bool               `json:"directAdd" bson:"directAdd"`
	Used                    int                `json:"used" bson:"used"`
	Belongs                 primitive.ObjectID `json:"belongs" bson:"belongs"`
}
