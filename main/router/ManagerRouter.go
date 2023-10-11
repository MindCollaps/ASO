package router

import (
	"ASO/main/database"
	"ASO/main/database/models"
	"ASO/main/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"text/template"
	"time"
)

func fetchAllGitUsers(c *gin.Context) ([]UserData, error) {
	cur, err := database.MongoDB.Collection("gitUser").Find(c, options.Find())
	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	var users []UserData
	for cur.Next(c) {
		var user models.GitHubUser
		cur.Decode(&user)

		//get userGroup
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id": user.UserGroup,
		}).Decode(&userGroup)

		if err != nil {
			continue
		}

		usrDt := UserData{
			ID:             user.ID.Hex(),
			GitHubUsername: user.GitHubUsername,
			GitHubID:       user.GitHubID,
			DateCreated:    user.DateCreated.Time().Format("01.02.2006"),
			DateExpires:    user.DateExpires.Time().Format("01.02.2006"),
			UserGroup:      userGroupModalToData(userGroup),
		}
		users = append(users, usrDt)
	}
	return users, nil
}

func userGroupModalToData(group models.UserGroup) UserGroupData {
	return UserGroupData{
		ID:          group.ID.Hex(),
		Name:        group.Name,
		Date:        group.Date.Time().Format("01.02.2006"),
		DateExpires: group.DateExpires.Time().Format("01.02.2006"),
	}
}

func fetchAllGroups(c *gin.Context) ([]UserGroupData, error) {
	cur, err := database.MongoDB.Collection("userGroup").Find(c, options.Find())
	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	var groups []UserGroupData
	for cur.Next(c) {
		var group models.UserGroup
		cur.Decode(&group)

		grp := userGroupModalToData(group)
		groups = append(groups, grp)
	}

	return groups, nil
}

type ManagerRouteData struct {
	Users  []UserData
	Groups []UserGroupData
}

type UserData struct {
	ID             string        `json:"id" bson:"_id"`
	GitHubUsername string        `json:"githubUsername" bson:"githubUsername"`
	GitHubID       string        `json:"githubID" bson:"githubID"`
	DateCreated    string        `json:"dateCreated" bson:"dateCreated"`
	DateExpires    string        `json:"dateExpires" bson:"dateExpires"`
	UserGroup      UserGroupData `json:"userGroup" bson:"userGroup"`
}

type UserGroupData struct {
	ID          string     `json:"id" bson:"_id"`
	Name        string     `json:"name" bson:"name"`
	Date        string     `json:"date" bson:"date"`
	DateExpires string     `json:"dateExpires" bson:"dateExpires"`
	Members     []UserData `json:"members" bson:"members"`
	Users       int        `json:"users" bson:"users"`
}

func initManagerRouter(router *gin.Engine) {
	router.GET("/manager", middleware.LoginToken(), func(c *gin.Context) {
		//display manager site

		users, err := fetchAllGitUsers(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching git users",
			})
			return
		}

		groups, err := fetchAllGroups(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching groups",
			})
			return
		}

		//for each group count how many users are assigned to that group
		for grp := range groups {
			count := 0
			for usr := range users {
				if users[usr].ID == groups[grp].ID {
					count++
				}
			}
			groups[grp].Users = count
		}

		data := ManagerRouteData{
			Groups: groups,
			Users:  users,
		}

		template := template.Must(template.ParseFiles("main/public/manager/index.gohtml"))
		template.Execute(c.Writer, data)
	})

	router.GET("/manager/user/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")

		var user models.GitHubUser
		err := database.MongoDB.Collection("gitUser").FindOne(c, bson.M{
			"_id": id,
		}).Decode(&user)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		//fetch user group from user
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id": user.UserGroup,
		}).Decode(&userGroup)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User group not found",
			})
			return
		}

		userData := UserData{
			ID:             user.ID.Hex(),
			GitHubUsername: user.GitHubUsername,
			GitHubID:       user.GitHubID,
			DateCreated:    user.DateCreated.Time().Format("01.02.2006"),
			DateExpires:    user.DateExpires.Time().Format("01.02.2006"),
			UserGroup: UserGroupData{
				ID:          userGroup.ID.Hex(),
				Name:        userGroup.Name,
				Date:        userGroup.Date.Time().Format("01.02.2006"),
				DateExpires: userGroup.DateExpires.Time().Format("01.02.2006"),
			},
		}

		template := template.Must(template.ParseFiles("main/public/manager/index.gohtml"))
		template.Execute(c.Writer, userData)
	})

	router.GET("/manager/group/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong user request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id": idd,
		}).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User group not found",
			})
			fmt.Println(err)
			return
		}

		//fetch members from group by fetchin all users and checking weather they are in the group or not
		var members []UserData

		users, err := fetchAllGitUsers(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching git users",
			})
			return
		}

		for i := range users {
			if users[i].UserGroup.ID == group.ID.Hex() {
				members = append(members, UserData{
					ID:             users[i].ID,
					GitHubUsername: users[i].GitHubUsername,
					GitHubID:       users[i].GitHubID,
					DateCreated:    users[i].DateCreated,
					DateExpires:    users[i].DateExpires,
				})
			}
		}

		groupData := userGroupModalToData(group)

		groupData.Members = members

		template := template.Must(template.ParseFiles("main/public/manager/group/index.gohtml"))
		template.Execute(c.Writer, groupData)
	})

	router.DELETE("manager/group/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong user request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		//delete group
		_, err = database.MongoDB.Collection("userGroup").DeleteOne(c, bson.M{
			"_id": idd,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when deleting group",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Group deleted",
		})
	})

	router.GET("/manager/group/create", middleware.LoginToken(), func(c *gin.Context) {
		template := template.Must(template.ParseFiles("main/public/manager/group/create/index.gohtml"))
		template.Execute(c.Writer, nil)
	})

	router.POST("/manager/group/create", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Name        string `json:"name" bson:"name"`
			DateExpires string `json:"dateExpires" bson:"dateExpires"`
		}
		//get from body
		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02", requestBody.DateExpires)

		//create group
		_, err = database.MongoDB.Collection("userGroup").InsertOne(c, models.UserGroup{
			ID:          primitive.NewObjectID(),
			Name:        requestBody.Name,
			Date:        primitive.NewDateTimeFromTime(time.Now()),
			DateExpires: primitive.NewDateTimeFromTime(dateExpiresTime),
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating group",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Group created",
		})
	})
}
