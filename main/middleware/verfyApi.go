package middleware

import (
	"ASO/main/crypt"
	"ASO/main/database"
	"ASO/main/database/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//check if the header has the api key

func LoginToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		//get jwt token from header
		token, err := c.Cookie("auth")

		if err != nil {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}
		if token == "" {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		jwt, err := crypt.ParseJwt(token)
		if err != nil {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		id, err := primitive.ObjectIDFromHex(jwt["userId"].(string))
		if err != nil {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		res := database.MongoDB.Collection("user").FindOne(c, bson.M{
			"_id": id,
		})

		if res.Err() != nil {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		//check time
		if jwt["exp"] != nil {
			if jwt["exp"].(float64) < jwt["iat"].(float64) {
				c.JSON(401, gin.H{
					"message": "Unauthorized",
				})
				c.Abort()
				return
			}
		}
		var dUser models.User
		err = res.Decode(&dUser)

		if err == nil {
			c.Set("user", dUser)
		}

		c.Set("userId", jwt["userId"])
		c.Set("loggedIn", true)
		c.Next()
	}
}
