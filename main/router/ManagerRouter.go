package router

import (
	"ASOServer/main/crypt"
	"ASOServer/main/database"
	"ASOServer/main/database/models"
	"ASOServer/main/git"
	"ASOServer/main/middleware"
	"ASOServer/main/tasks"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/softbrewery/gojoi/pkg/joi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"text/template"
	"time"
)

func fetchAllGitUsers(c *gin.Context) ([]GitUserData, error) {
	//check if admin
	isSuperUser := c.MustGet("user").(models.User).IsSuperUser

	belongsBson := bson.M{
		"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
	}

	if isSuperUser {
		belongsBson = bson.M{}
	}

	cur, err := database.MongoDB.Collection("gitUser").Find(c, belongsBson)
	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	var users []GitUserData
	for cur.Next(c) {
		var user models.GitHubUser
		cur.Decode(&user)

		//get userGroup
		var userGroup models.UserGroup
		addedBson := bson.M{
			"_id": user.UserGroup,
		}
		if !isSuperUser {
			addedBson["belongs"] = belongsBson["belongs"]
		}
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&userGroup)

		var usrDt GitUserData

		expired := time.Now().After(user.DateExpires.Time())

		if err != nil {
			usrDt = GitUserData{
				ID:             user.ID.Hex(),
				GitHubUsername: user.GitHubUsername,
				DateCreated:    user.DateCreated.Time().Format("2006-01-02 15:04:05"),
				DateExpires:    user.DateExpires.Time().Format("2006-01-02 15:04:05"),
				Username:       user.Username,
				IsExpired:      expired,
				ExpiryByGroup:  user.ExpiresGroup,
				AddedToRepo:    user.AddedToRepo,
				Expires:        user.Expires,
			}
		} else {
			usrDt = GitUserData{
				ID:             user.ID.Hex(),
				GitHubUsername: user.GitHubUsername,
				DateCreated:    user.DateCreated.Time().Format("2006-01-02 15:04:05"),
				DateExpires:    user.DateExpires.Time().Format("2006-01-02 15:04:05"),
				UserGroup:      userGroupModalToData(userGroup),
				Username:       user.Username,
				IsExpired:      expired,
				ExpiryByGroup:  user.ExpiresGroup,
				AddedToRepo:    user.AddedToRepo,
				Expires:        user.Expires,
			}
		}
		if usrDt.DateExpires == "0001-01-01 01:00:00" {
			usrDt.DateExpires = "Never"
		}
		users = append(users, usrDt)
	}
	return users, nil
}

func userGroupModalToData(group models.UserGroup) UserGroupData {
	expired := time.Now().After(group.DateExpires.Time())
	var g = UserGroupData{
		ID:              group.ID.Hex(),
		Name:            group.Name,
		Date:            group.Date.Time().Format("2006-01-02 15:04:05"),
		DateExpires:     group.DateExpires.Time().Format("2006-01-02 15:04:05"),
		Expires:         group.Expires,
		AutoDelete:      group.AutoDelete,
		Notify:          group.Notify,
		GitHubRepo:      group.GitHubRepo,
		GitHubOwner:     group.GitHubOwner,
		IsExpired:       expired,
		AutoRemoveUsers: group.AutoRemoveUsers,
	}

	if g.DateExpires == "0001-01-01 01:00:00" {
		g.DateExpires = "Never"
	}
	return g
}

func tokenModalToData(token models.Token) TokenData {
	expired := time.Now().After(token.DateExpires.Time())
	return TokenData{
		ID:          token.ID.Hex(),
		Name:        token.Name,
		Count:       token.Count,
		Token:       token.Token,
		DateCreated: token.DateCreated.Time().Format("2006-01-02 15:04:05"),
		DateExpires: token.DateExpires.Time().Format("2006-01-02 15:04:05"),
		DirectAdd:   token.DirectAdd,
		Used:        token.Used,
		IsExpired:   expired,
	}
}

func notificationModalToData(notification models.Notification) NotificationData {
	dateCreated := notification.DateCreated.Time().Format("2006-01-02 15:04:05")
	data := NotificationData{
		ID:           notification.ID.Hex(),
		Notification: notification.Notification,
		DateCreated:  dateCreated,
		Title:        notification.Title,
		Style:        notification.Style,
		Seen:         notification.Seen,
		Popup:        notification.Popup,
		Profile:      notification.Profile,
	}

	if notification.UserGroup != primitive.NilObjectID {
		data.UserGroup = notification.UserGroup.Hex()
	}

	if notification.GitHubUser != primitive.NilObjectID {
		data.GitHubUser = notification.GitHubUser.Hex()
	}

	if notification.Token != primitive.NilObjectID {
		data.Token = notification.Token.Hex()
	}

	return data
}

func fetchAllGroups(c *gin.Context) ([]UserGroupData, error) {
	belonBson := bson.M{
		"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
	}

	if c.MustGet("user").(models.User).IsSuperUser {
		belonBson = bson.M{}
	}

	cur, err := database.MongoDB.Collection("userGroup").Find(c, belonBson)
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

func fetchAllTokens(c *gin.Context) ([]TokenData, error) {
	belonBson := bson.M{
		"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
	}

	if c.MustGet("user").(models.User).IsSuperUser {
		belonBson = bson.M{}
	}

	cur, err := database.MongoDB.Collection("token").Find(c, belonBson)
	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	var tokens []TokenData
	for cur.Next(c) {
		var token models.Token
		cur.Decode(&token)

		//get userGroup
		var userGroup models.UserGroup
		addedBson := bson.M{
			"_id": token.UserGroup,
		}
		if !c.MustGet("user").(models.User).IsSuperUser {
			addedBson["belongs"] = belonBson["belongs"]
		}
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&userGroup)

		grpErr := false
		if err != nil {
			grpErr = true
		}

		//get createdBy
		var createdBy models.User
		addedBson = bson.M{
			"_id": token.CreatedBy,
		}

		if !c.MustGet("user").(models.User).IsSuperUser {
			addedBson["belongs"] = belonBson["belongs"]
		}
		err = database.MongoDB.Collection("gitusr").FindOne(c, addedBson).Decode(&createdBy)

		createdByErr := false
		if err != nil {
			createdByErr = true
		}

		tkn := tokenModalToData(token)
		if !grpErr {
			tkn.UserGroup = userGroupModalToData(userGroup)
		}
		if !createdByErr {
			tkn.CreatedBy = createdBy.Username
		}

		tokens = append(tokens, tkn)
	}

	return tokens, nil
}

func fetchAllUsers(c *gin.Context) ([]UserData, error) {
	cur, err := database.MongoDB.Collection("user").Find(c, options.Find())
	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	var users []UserData
	for cur.Next(c) {
		var user models.User
		cur.Decode(&user)

		usr := userModalToData(user)
		users = append(users, usr)
	}

	return users, nil
}

func fetchAllNotifications(c *gin.Context) ([]NotificationData, error) {
	cur, err := database.MongoDB.Collection("notification").Find(c, bson.M{
		"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
	})

	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	var notifications []NotificationData
	for cur.Next(c) {
		var notification models.Notification
		cur.Decode(&notification)

		ntf := notificationModalToData(notification)

		notifications = append(notifications, ntf)
	}

	return notifications, nil
}

func fetchGroupById(c *gin.Context, id string) (UserGroupData, error) {
	idd, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return UserGroupData{}, err
	}

	var group models.UserGroup
	err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
		"_id": idd,
	}).Decode(&group)

	if err != nil {
		return UserGroupData{}, err
	}

	return userGroupModalToData(group), nil
}

func userModalToData(user models.User) UserData {
	dateCreated := user.DateCreated.Time().Format("2006-01-02 15:04:05")

	return UserData{
		ID:             user.ID.Hex(),
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		DateCreated:    dateCreated,
		IsSuperUser:    user.IsSuperUser,
		Token:          user.GitHubToken,
	}

}

type ManagerRouteData struct {
	GUsers        []GitUserData
	Users         []UserData
	Groups        []UserGroupData
	Tokens        []TokenData
	SuperUser     bool
	User          UserData
	Notifications []NotificationData
}

type ManagerCreateTkData struct {
	Groups []UserGroupData
}

type ManagerTokenData struct {
	Token string `json:"token" bson:"token"`
}

type ManagerTkData struct {
	Failed  bool
	Message string
}

type GitUserData struct {
	ID             string          `json:"id" bson:"_id"`
	Username       string          `json:"username" bson:"username"`
	GitHubUsername string          `json:"githubUsername" bson:"githubUsername"`
	GitHubID       string          `json:"githubID" bson:"githubID"`
	DateCreated    string          `json:"dateCreated" bson:"dateCreated"`
	DateExpires    string          `json:"dateExpires" bson:"dateExpires"`
	ExpiryByGroup  bool            `json:"expiryByGroup" bson:"expiryByGroup"`
	UserGroup      UserGroupData   `json:"userGroup" bson:"userGroup"`
	Groups         []UserGroupData `json:"groups" bson:"groups"`
	Expires        bool            `json:"expires" bson:"expires"`
	IsCollaborator bool            `json:"isCollaborator" bson:"isCollaborator"`
	IsInvited      bool            `json:"isInvited" bson:"isInvited"`
	IsExpired      bool            `json:"isExpired" bson:"isExpired"`
	AddedToRepo    bool            `json:"addedToRepo" bson:"addedToRepo"`
}

type UserGroupData struct {
	ID              string        `json:"id" bson:"_id"`
	Name            string        `json:"name" bson:"name"`
	Date            string        `json:"date" bson:"date"`
	DateExpires     string        `json:"dateExpires" bson:"dateExpires"`
	Members         []GitUserData `json:"members" bson:"members"`
	Users           int           `json:"users" bson:"users"`
	Expires         bool          `json:"expires" bson:"expires"`
	AutoDelete      bool          `json:"autoDelete" bson:"autoDelete"`
	Notify          bool          `json:"notify" bson:"notify"`
	GitHubRepo      string        `json:"githubRepo" bson:"githubRepo"`
	GitHubOwner     string        `json:"githubOwner" bson:"githubOwner"`
	IsExpired       bool          `json:"isExpired" bson:"isExpired"`
	AutoRemoveUsers bool          `json:"autoRemoveUsers" bson:"autoRemoveUsers"`
}

type TokenData struct {
	ID          string        `json:"id" bson:"_id"`
	Name        string        `json:"name" bson:"name"`
	Count       int           `json:"count" bson:"count"`
	Token       string        `json:"token" bson:"token"`
	UserGroup   UserGroupData `json:"userGroup" bson:"userGroup"`
	DateCreated string        `json:"dateCreated" bson:"dateCreated"`
	DateExpires string        `json:"dateExpires" bson:"dateExpires"`
	DirectAdd   bool          `json:"directAdd" bson:"directAdd"`
	AutoDelete  bool          `json:"autoDelete" bson:"autoDelete"`
	Notify      bool          `json:"notify" bson:"notify"`
	Used        int           `json:"used" bson:"used"`
	CreatedBy   string        `json:"createdBy" bson:"createdBy"`
	IsExpired   bool          `json:"isExpired" bson:"isExpired"`
}

type UserData struct {
	ID             string `json:"id" bson:"_id"`
	Username       string `json:"username" bson:"username"`
	Email          string `json:"email" bson:"email"`
	GitHubUsername string `json:"githubUsername" bson:"githubUsername"`
	DateCreated    string `json:"dateCreated" bson:"dateCreated"`
	IsSuperUser    bool   `json:"isSuperUser" bson:"isSuperUser"`
	Token          string `json:"token" bson:"token"`
}

type NotificationData struct {
	ID           string `json:"id" bson:"_id"`
	Notification string `json:"notification" bson:"notification"`
	DateCreated  string `json:"dateCreated" bson:"dateCreated"`
	Title        string `json:"title" bson:"title"`
	Style        string `json:"style" bson:"style"`
	UserGroup    string `json:"userGroup" bson:"userGroup"`
	GitHubUser   string `json:"githubUser" bson:"githubUser"`
	Token        string `json:"token" bson:"token"`
	Seen         bool   `json:"seen" bson:"seen"`
	Popup        bool   `json:"popup" bson:"popup"`
	Profile      bool   `json:"profile" bson:"profile"`
}

func initManagerRouter(router *gin.Engine) {
	router.GET("/manager", middleware.LoginToken(), func(c *gin.Context) {
		//display manager site

		gitUsers, err := fetchAllGitUsers(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching git gitUsers",
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

		//for each group count how many gitUsers are assigned to that group
		for grp := range groups {
			count := 0
			for usr := range gitUsers {
				if gitUsers[usr].UserGroup.ID == groups[grp].ID {
					count++
				}
			}
			groups[grp].Users = count
		}

		tokens, err := fetchAllTokens(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching tokens",
			})
			return
		}

		isSuperUser := c.MustGet("user").(models.User).IsSuperUser

		users, err := fetchAllUsers(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching gitUsers",
			})
			return
		}

		notifications, err := fetchAllNotifications(c)

		data := ManagerRouteData{
			Groups:        groups,
			GUsers:        gitUsers,
			Tokens:        tokens,
			SuperUser:     isSuperUser,
			Users:         users,
			User:          userModalToData(c.MustGet("user").(models.User)),
			Notifications: notifications,
		}

		template := template.Must(template.ParseFS(files, "main/public/manager/index.gohtml"))
		template.Execute(c.Writer, data)
	})

	router.GET("/manager/gitusr/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")

		usr := c.MustGet("user").(models.User)

		//check if id is valid
		idd, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitusr id",
			})
			return
		}

		var user models.GitHubUser
		addedBson := bson.M{
			"_id": idd,
		}

		if !usr.IsSuperUser {
			addedBson["belongs"] = c.MustGet("userIdPrimitive").(primitive.ObjectID)
		}
		err = database.MongoDB.Collection("gitUser").FindOne(c, addedBson).Decode(&user)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		//fetch gitusr group from gitusr
		addedBson = bson.M{
			"_id": user.UserGroup,
		}

		if !usr.IsSuperUser {
			addedBson["belongs"] = c.MustGet("userIdPrimitive").(primitive.ObjectID)
		}
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&userGroup)

		//fetch all groups
		grps, err := fetchAllGroups(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching groups",
			})
			return
		}

		var userData GitUserData

		if userGroup.ID == primitive.NilObjectID {
			userData = GitUserData{
				ID:             user.ID.Hex(),
				GitHubUsername: user.GitHubUsername,
				DateCreated:    user.DateCreated.Time().Format("2006-01-02 15:04:05"),
				DateExpires:    user.DateExpires.Time().Format("2006-01-02 15:04:05"),
				ExpiryByGroup:  user.ExpiresGroup,
				Groups:         grps,
				Username:       user.Username,
			}
		} else {
			userGrpData := userGroupModalToData(userGroup)
			userData = GitUserData{
				ID:             user.ID.Hex(),
				GitHubUsername: user.GitHubUsername,
				Username:       user.Username,
				ExpiryByGroup:  user.ExpiresGroup,
				DateCreated:    user.DateCreated.Time().Format("2006-01-02 15:04:05"),
				DateExpires:    user.DateExpires.Time().Format("2006-01-02 15:04:05"),
				UserGroup:      userGrpData,
				Groups:         grps,
			}
		}

		if userData.DateExpires == "0001-01-01 01:00:00" {
			userData.DateExpires = "Never"
		}

		if userGroup.ID != primitive.NilObjectID {
			userData.IsCollaborator = user.AddedToRepo
			if userData.IsCollaborator {
				userData.IsInvited = false
			} else {
				userData.IsInvited = git.CheckIfUserIsPendingInvite(userGroup.GitHubOwner, user.GitHubUsername, usr.GitHubToken, userGroup.GitHubRepo)
				if !userData.IsInvited {
					userData.IsCollaborator = git.CheckIfUserIsColabo(userGroup.GitHubOwner, user.GitHubUsername, usr.GitHubToken, userGroup.GitHubRepo)

					if userData.IsCollaborator {
						//add user to group
						_, err = database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
							"_id": user.ID,
						}, bson.M{
							"$set": bson.M{
								"addedToRepo": true,
							},
						})

						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		} else {
			userData.IsCollaborator = false
			userData.IsInvited = false
		}

		template := template.Must(template.ParseFS(files, "main/public/manager/gitusr/index.gohtml"))
		template.Execute(c.Writer, userData)
	})

	router.GET("/manager/group/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid gitusr id",
			})
			return
		}

		usr := c.MustGet("user").(models.User)

		addedBson := bson.M{
			"_id": idd,
		}

		if !usr.IsSuperUser {
			addedBson["belongs"] = c.MustGet("userIdPrimitive").(primitive.ObjectID)
		}
		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User group not found",
			})
			log.Println(err)
			return
		}

		//fetch members from group by fetchin all users and checking weather they are in the group or not
		var members []GitUserData

		users, err := fetchAllGitUsers(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching git users",
			})
			return
		}

		for i := range users {
			if users[i].UserGroup.ID == group.ID.Hex() {
				user := users[i]

				if group.ID != primitive.NilObjectID {
					user.IsCollaborator = user.AddedToRepo
					if user.IsCollaborator {
						user.IsInvited = false
					} else {
						user.IsInvited = git.CheckIfUserIsPendingInvite(group.GitHubOwner, user.GitHubUsername, usr.GitHubToken, group.GitHubRepo)
						if !user.IsInvited {
							user.IsCollaborator = git.CheckIfUserIsColabo(group.GitHubOwner, user.GitHubUsername, usr.GitHubToken, group.GitHubRepo)

							if user.IsCollaborator {
								//add user to group
								_, err = database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
									"_id": users[i].ID,
								}, bson.M{
									"$set": bson.M{
										"addedToRepo":    true,
										"isCollaborator": true,
									},
								})

								if err != nil {
									log.Println(err)
								}
							}
						}
					}
				} else {
					user.IsCollaborator = false
					user.IsInvited = false
				}

				members = append(members, user)
			}
		}

		groupData := userGroupModalToData(group)

		if len(members) > 0 {
			groupData.Members = members
		}

		template := template.Must(template.ParseFS(files, "main/public/manager/group/index.gohtml"))
		template.Execute(c.Writer, groupData)
	})

	router.DELETE("manager/group/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid gitusr id",
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
		template := template.Must(template.ParseFS(files, "main/public/manager/group/create/index.gohtml"))
		template.Execute(c.Writer, nil)
	})

	router.POST("/manager/group/create", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Name            string `json:"name" bson:"name" binding:"required"`
			DateExpires     string `json:"dateExpires" bson:"dateExpires"`
			Notify          bool   `json:"notify" bson:"notify"`
			Expires         bool   `json:"doesExpire" bson:"doesExpire"`
			AutoDelete      bool   `json:"autoDelete" bson:"autoDelete"`
			GitRepo         string `json:"gitRepo" bson:"gitRepo" binding:"required"`
			GitOwner        string `json:"gitOwner" bson:"gitOwner"`
			IsOwn           bool   `json:"isOwn" bson:"isOwn"`
			AutoRemoveUsers bool   `json:"autoRemoveUsers" bson:"autoRemoveUsers"`
		}

		//get from body
		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04:05.000Z", requestBody.DateExpires)

		if requestBody.Expires {
			if err != nil || dateExpiresTime.Before(time.Now()) {
				if err != nil {
					log.Println(err)
					c.JSON(400, gin.H{
						"message": "Invalid date",
					})
				} else {
					log.Println("Date must be in the future")
					c.JSON(400, gin.H{
						"message": "Date must be in the future",
					})
				}
				return
			}
		}

		if !requestBody.Expires {
			//set time to 0
			dateExpiresTime = time.Time{}
		}

		//check repo
		var owner = ""

		if requestBody.IsOwn {
			owner = c.MustGet("user").(models.User).GitHubUsername
		} else {
			owner = requestBody.GitOwner
		}

		if !git.CheckRepoExists(owner, c.MustGet("user").(models.User).GitHubToken, requestBody.GitRepo) {
			c.JSON(400, gin.H{
				"message": "Invalid git repo",
			})
			return
		}

		//create group
		_, err = database.MongoDB.Collection("userGroup").InsertOne(c, models.UserGroup{
			ID:              primitive.NewObjectID(),
			Name:            requestBody.Name,
			Date:            primitive.NewDateTimeFromTime(time.Now()),
			DateExpires:     primitive.NewDateTimeFromTime(dateExpiresTime),
			AutoDelete:      requestBody.AutoDelete,
			Notify:          requestBody.Notify,
			NotifiedExpired: false,
			NotifiedDeleted: false,
			Expires:         requestBody.Expires,
			Belongs:         c.MustGet("userIdPrimitive").(primitive.ObjectID),
			GitHubRepo:      requestBody.GitRepo,
			GitHubOwner:     owner,
			AutoRemoveUsers: requestBody.AutoRemoveUsers,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating group",
			})
			log.Println(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "Group created",
		})
	})

	router.GET("/manager/group/edit/:id", middleware.LoginToken(), func(c *gin.Context) {
		grp, err := fetchGroupById(c, c.Param("id"))
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching group",
			})
			return
		}

		template := template.Must(template.ParseFS(files, "main/public/manager/group/edit/index.gohtml"))
		template.Execute(c.Writer, grp)
	})

	router.POST("/manager/group/edit/:id", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Name            string `json:"name" bson:"name" binding:"required"`
			DateExpires     string `json:"dateExpires" bson:"dateExpires"`
			Notify          bool   `json:"notify" bson:"notify"`
			Expires         bool   `json:"doesExpire" bson:"doesExpire"`
			AutoDelete      bool   `json:"autoDelete" bson:"autoDelete"`
			GitRepo         string `json:"gitRepo" bson:"gitRepo" binding:"required"`
			GitOwner        string `json:"gitOwner" bson:"gitOwner"`
			AutoRemoveUsers bool   `json:"autoRemoveUsers" bson:"autoRemoveUsers"`
		}

		//check id
		idd, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitusr group id",
			})
			log.Println(err)
			return
		}

		//get from body
		err = c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04:05.000Z", requestBody.DateExpires)

		if requestBody.Expires {
			if err != nil || dateExpiresTime.Before(time.Now()) {
				if err != nil {
					log.Println(err)
					c.JSON(400, gin.H{
						"message": "Invalid date",
					})
				} else {
					log.Println("Date must be in the future")
					c.JSON(400, gin.H{
						"message": "Date must be in the future",
					})
				}
				return
			}
		}

		if !requestBody.Expires {
			//set time to 0
			dateExpiresTime = time.Time{}
		}

		//create group
		_, err = database.MongoDB.Collection("userGroup").UpdateOne(c, gin.H{
			"_id": idd,
		},
			gin.H{
				"$set": gin.H{
					"name":            requestBody.Name,
					"dateExpires":     primitive.NewDateTimeFromTime(dateExpiresTime),
					"autoDelete":      requestBody.AutoDelete,
					"notify":          requestBody.Notify,
					"expires":         requestBody.Expires,
					"autoRemoveUsers": requestBody.AutoRemoveUsers,
					"gitHubRepo":      requestBody.GitRepo,
					"gitHubOwner":     requestBody.GitOwner,
				},
			})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when updating group",
			})
			log.Println(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "Group updated",
		})
	})

	router.GET("manager/groups/refresh", middleware.LoginToken(), func(c *gin.Context) {
		tasks.RefreshGitState()
	})

	router.GET("/manager/token/create", middleware.LoginToken(), func(c *gin.Context) {
		grps, err := fetchAllGroups(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching groups",
			})
			return
		}

		template := template.Must(template.ParseFS(files, "main/public/manager/token/create/index.gohtml"))
		template.Execute(c.Writer, ManagerCreateTkData{
			Groups: grps,
		})
	})

	router.POST("/manager/token/create", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Name        string `json:"name" bson:"name" binding:"required"`
			UserGroup   string `json:"userGroup" bson:"userGroup" binding:"required"`
			DirectAdd   bool   `json:"directAdd" bson:"directAdd"`
			AutoDelete  bool   `json:"autoDelete" bson:"autoDelete"`
			Notify      bool   `json:"notify" bson:"notify"`
			DateExpires string `json:"dateExpires" bson:"dateExpires" binding:"required"`
			Count       string `json:"count" bson:"count" binding:"required"`
		}

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		//check UserGroup
		grp, err := primitive.ObjectIDFromHex(requestBody.UserGroup)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitusr group id",
			})
			log.Println(err)
			return
		}

		//check if exists
		addedBson := bson.M{
			"_id": grp,
		}

		if !c.MustGet("user").(models.User).IsSuperUser {
			addedBson["belongs"] = c.MustGet("userIdPrimitive").(primitive.ObjectID)
		}

		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&userGroup)

		//parse count
		count, err := strconv.Atoi(requestBody.Count)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid count",
			})
			log.Println(err)
			return
		}
		//check count
		if count < 1 {
			c.JSON(400, gin.H{
				"message": "Invalid count",
			})

			log.Println("Invalid count")
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04:05.000Z", requestBody.DateExpires)

		if err != nil || dateExpiresTime.Before(time.Now()) {
			if err != nil {
				log.Println(err)
				c.JSON(400, gin.H{
					"message": "Invalid date",
				})
			} else {
				log.Println("Date must be in the future")
				c.JSON(400, gin.H{
					"message": "Date must be in the future",
				})
			}
			return
		}

		token := GenerateRandomString(6)

		//create token
		_, err = database.MongoDB.Collection("token").InsertOne(c, models.Token{
			ID:                      primitive.NewObjectID(),
			Name:                    requestBody.Name,
			Count:                   count,
			Token:                   token,
			UserGroup:               userGroup.ID,
			DateCreated:             primitive.NewDateTimeFromTime(time.Now()),
			DateExpires:             primitive.NewDateTimeFromTime(dateExpiresTime),
			CreatedBy:               c.MustGet("userIdPrimitive").(primitive.ObjectID),
			IsUserRegistrationToken: false,
			DirectAdd:               requestBody.DirectAdd,
			Used:                    0,
			Belongs:                 c.MustGet("userIdPrimitive").(primitive.ObjectID),
			SuperUser:               false,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating token",
			})
			log.Println(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "Token created",
			"token":   token,
		})
	})

	router.GET("/token/:tk", middleware.LoginToken(), func(c *gin.Context) {
		token := c.Param("tk")

		//check if token exists
		addedBson := bson.M{
			"token": token,
		}

		if !c.MustGet("user").(models.User).IsSuperUser {
			addedBson["belongs"] = c.MustGet("userIdPrimitive").(primitive.ObjectID)
		}
		var tk models.Token
		err := database.MongoDB.Collection("token").FindOne(c, addedBson).Decode(&tk)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Token not found",
			})
			return
		}

		template := template.Must(template.ParseFS(files, "main/public/token/index.gohtml"))
		template.Execute(c.Writer, ManagerTokenData{
			Token: token,
		})
	})

	router.DELETE("/manager/token/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid token id",
			})
			return
		}

		//delete token
		_, err = database.MongoDB.Collection("token").DeleteOne(c, bson.M{
			"_id": idd,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when deleting token",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Token deleted",
		})
	})

	router.GET("/manager/token/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid token id",
			})
			return
		}

		//find token
		addedBson := bson.M{
			"_id": idd,
		}

		if !c.MustGet("user").(models.User).IsSuperUser {
			addedBson["belongs"] = c.MustGet("userIdPrimitive").(primitive.ObjectID)
		}

		var token models.Token
		err = database.MongoDB.Collection("token").FindOne(c, addedBson).Decode(&token)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Token not found",
			})
			return
		}

		tk := tokenModalToData(token)

		template := template.Must(template.ParseFS(files, "main/public/manager/token/index.gohtml"))
		template.Execute(c.Writer, tk)
	})

	router.GET("/tk/:tk", func(c *gin.Context) {
		token := c.Param("tk")

		//check if token exists
		var tk models.Token
		err := database.MongoDB.Collection("token").FindOne(c, bson.M{
			"token": token,
		}).Decode(&tk)

		failed := false
		message := "no message"

		if err != nil {
			failed = true
			message = "Token not found"
		} else if tk.DateExpires.Time().Before(time.Now()) {
			failed = true
			message = "Token is already expired"
		} else if tk.Used >= tk.Count {
			failed = true
			message = "Tokens usage limit reached"
		} else if tk.IsUserRegistrationToken {
			jwt, err := crypt.GenerateRegToken(token)
			if err != nil {
				failed = true
				message = "Internal server error when generating jwt"
			} else {
				c.SetCookie("regauth", jwt, 3600, "/", "", false, true)
				c.Redirect(302, "/reg")
				return
			}
		}

		template := template.Must(template.ParseFS(files, "main/public/tk/index.gohtml", "main/public/templates/template.gohtml"))
		template.Execute(c.Writer, ManagerTkData{
			Failed:  failed,
			Message: message,
		})
	})

	router.POST("/tk", func(c *gin.Context) {
		var requestBody struct {
			Token          string `json:"token" bson:"token" binding:"required"`
			GitHubUsername string `json:"gitUsername" bson:"gitUsername" binding:"required"`
			Username       string `json:"username" bson:"username" binding:"required"`
		}

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		//joi validation
		if err := joi.Validate(requestBody.Username, UsernameSchema); err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid username",
				"error":   err.Error(),
				"field":   "username",
			})

			log.Println("Invalid username")
			return
		}

		if err := joi.Validate(requestBody.GitHubUsername, GitHubUsername); err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid github username",
				"error":   err.Error(),
				"field":   "gitUsername",
			})

			log.Println("Invalid github username")
			return
		}

		//check if the username is okay length, characters etc
		username := requestBody.Username

		//check if token exists

		var tk models.Token
		err = database.MongoDB.Collection("token").FindOne(c, bson.M{
			"token": requestBody.Token,
		}).Decode(&tk)

		if err != nil {
			c.JSON(404, gin.H{
				"message":    "Token not found",
				"tokenError": true,
			})

			log.Println("Token not found")
			return
		}

		if tk.IsUserRegistrationToken {
			c.JSON(400, gin.H{
				"message":    "Token is a user registration token",
				"tokenError": true,
			})
			log.Println("Token is a user registration token")
			return
		}

		//check if token is expired
		if tk.DateExpires.Time().Before(time.Now()) {
			c.JSON(400, gin.H{
				"message":    "Token is already expired",
				"tokenError": true,
			})
			log.Println("Token is already expired")
			return
		}

		//check if token is used
		if tk.Used >= tk.Count {
			c.JSON(400, gin.H{
				"message":    "Tokens usage limit reached",
				"tokenError": true,
			})
			log.Println("Tokens usage limit reached")
			return
		}

		var user models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, bson.M{
			"githubUsername": requestBody.GitHubUsername,
			"userGroup":      tk.UserGroup,
		}).Decode(&user)

		if err == nil {
			c.JSON(400, gin.H{
				"message": "User already exists in this group",
			})
			log.Println("User already exists")
			return
		}

		//get userGroup
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id": tk.UserGroup,
		}).Decode(&userGroup)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User group not found",
			})
			log.Println("User group not found")
			return
		}

		//get creator
		var creator models.User
		err = database.MongoDB.Collection("user").FindOne(c, bson.M{
			"_id": tk.CreatedBy,
		}).Decode(&creator)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Creator not found",
			})
			log.Println("Creator not found")
			return
		}

		if !git.CheckUser(requestBody.GitHubUsername, creator.GitHubToken) {
			c.JSON(400, gin.H{
				"message": "User can not be found on github",
			})
			log.Println("User can not be found on github")
			return
		}

		isColabo := false

		if tk.DirectAdd {
			if !git.AddUserToRepo(requestBody.GitHubUsername, creator.GitHubToken, userGroup.GitHubRepo, userGroup.GitHubOwner) {
				c.JSON(500, gin.H{
					"message": "Internal server error when adding gitusr to repo",
				})
				log.Println("Internal server error when adding gitusr to repo")
				return
			}
		} else {
			isColabo = git.CheckIfUserIsColabo(userGroup.GitHubOwner, requestBody.GitHubUsername, creator.GitHubToken, userGroup.GitHubRepo)
		}

		//create gitusr
		_, err = database.MongoDB.Collection("gitUser").InsertOne(c, models.GitHubUser{
			ID:             primitive.NewObjectID(),
			Username:       username,
			GitHubUsername: requestBody.GitHubUsername,
			DateCreated:    primitive.NewDateTimeFromTime(time.Now()),
			Expires:        true,
			ExpiresGroup:   true,
			DateExpires:    userGroup.DateExpires,
			UserGroup:      userGroup.ID,
			AddedToRepo:    isColabo,
			Belongs:        creator.ID,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating gitusr",
			})
			log.Println("Internal server error when creating gitusr")
			return
		}

		tk.Used++

		//update token
		if tk.Used >= tk.Count {
			//delete
			_, err = database.MongoDB.Collection("token").DeleteOne(c, bson.M{
				"_id": tk.ID,
			})

			database.MongoDB.Collection("notification").InsertOne(c, models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      creator.ID,
				Notification: "Token " + tk.Name + " has reached it's usage limit and was therefore deleted!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Token " + tk.Name + " deleted",
				UserGroup:    primitive.NilObjectID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "info",
			})
		} else {
			_, err = database.MongoDB.Collection("token").UpdateOne(c, bson.M{
				"_id": tk.ID,
			}, bson.M{
				"$set": bson.M{
					"used": tk.Used,
				},
			})
		}

		c.JSON(200, gin.H{
			"message": "User created",
		})
	})

	router.GET("/manager/gitusr/create", middleware.LoginToken(), func(c *gin.Context) {
		grps, err := fetchAllGroups(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching groups",
			})
			return
		}

		template := template.Must(template.ParseFS(files, "main/public/manager/gitusr/create/index.gohtml"))
		template.Execute(c.Writer, ManagerCreateTkData{
			Groups: grps,
		})
	})

	router.POST("/manager/gitusr/create", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Username    string `json:"username" bson:"username" binding:"required"`
			GitUsername string `json:"gitUsername" bson:"gitUsername" binding:"required"`
			UserGroup   string `json:"userGroup" bson:"userGroup" binding:"required"`
			ExpireGroup bool   `json:"expireGroup" bson:"expireGroup"`
			Expires     bool   `json:"expires" bson:"expires"`
			DateExpires string `json:"dateExpires" bson:"dateExpires"`
		}

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		//check if gitusr group exists
		grp, err := primitive.ObjectIDFromHex(requestBody.UserGroup)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitusr group id",
			})
			log.Println(err)
			return
		}

		//check if gitusr group exists
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id":     grp,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&userGroup)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User group not found",
			})
			log.Println("User group not found")
			return
		}

		//check if the username is okay length, characters etc
		username := requestBody.Username
		if len(username) < 3 {
			c.JSON(400, gin.H{
				"message": "Username is too short",
			})
			log.Println("Username is too short")
			return
		}

		if len(username) > 20 {
			c.JSON(400, gin.H{
				"message": "Username is too long",
			})
			log.Println("Username is too long")
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04:05.000Z", requestBody.DateExpires)

		if requestBody.Expires && !requestBody.ExpireGroup {
			if err != nil || dateExpiresTime.Before(time.Now()) {
				c.JSON(400, gin.H{
					"message": "Invalid date",
				})
				if err != nil {
					log.Println(err)
				} else {
					log.Println("Invalid date")
				}
				return
			}
		}

		if !requestBody.Expires && !requestBody.ExpireGroup {
			//set time to 0
			dateExpiresTime = time.Time{}
		}

		//create gitusr
		_, err = database.MongoDB.Collection("gitUser").InsertOne(c, models.GitHubUser{
			ID:             primitive.NewObjectID(),
			Username:       username,
			GitHubUsername: requestBody.GitUsername,
			DateCreated:    primitive.NewDateTimeFromTime(time.Now()),
			Expires:        requestBody.Expires,
			ExpiresGroup:   requestBody.ExpireGroup,
			DateExpires:    primitive.NewDateTimeFromTime(dateExpiresTime),
			UserGroup:      userGroup.ID,
			AddedToRepo:    false,
			Belongs:        c.MustGet("userIdPrimitive").(primitive.ObjectID),
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating gitusr",
			})
			log.Println("Internal server error when creating gitusr")
			return
		}

		c.JSON(200, gin.H{
			"message": "User created",
		})
	})

	router.DELETE("/manager/gitusr/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid gitusr id",
			})
			return
		}

		//delete gitusr
		_, err = database.MongoDB.Collection("gitUser").DeleteOne(c, bson.M{
			"_id": idd,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when deleting gitusr",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Gitusr deleted",
		})
	})

	router.GET("/manager/user/:id", middleware.SuperUser(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		//find user

		var user models.User
		err = database.MongoDB.Collection("user").FindOne(c, bson.M{
			"_id": idd,
		}).Decode(&user)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		usr := userModalToData(user)

		template := template.Must(template.ParseFS(files, "main/public/manager/user/index.gohtml"))
		template.Execute(c.Writer, usr)
	})

	router.DELETE("/manager/user/:id", middleware.SuperUser(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		//delete user
		_, err = database.MongoDB.Collection("user").DeleteOne(c, bson.M{
			"_id": idd,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when deleting user",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "User deleted",
		})
	})

	router.GET("/manager/user/create", middleware.SuperUser(), func(c *gin.Context) {
		template := template.Must(template.ParseFS(files, "main/public/manager/user/create/index.gohtml"))
		template.Execute(c.Writer, nil)
	})

	router.POST("/manager/user/create", middleware.SuperUser(), func(c *gin.Context) {
		var requestBody struct {
			SuperUser   bool   `json:"superUser" bson:"superUser"`
			DateExpires string `json:"dateExpires" bson:"dateExpires""`
			Name        string `json:"name" bson:"name" binding:"required"`
			Count       string `json:"count" bson:"count" binding:"required"`
		}

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04:05.000Z", requestBody.DateExpires)

		if err != nil || dateExpiresTime.Before(time.Now()) {
			if err != nil {
				log.Println(err)
				c.JSON(400, gin.H{
					"message": "Invalid date",
				})
			} else {
				log.Println("Date must be in the future")
				c.JSON(400, gin.H{
					"message": "Date must be in the future",
				})
			}
			return
		}

		//create token
		token := GenerateRandomString(6)

		//parse count
		count, err := strconv.Atoi(requestBody.Count)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid count",
			})
			log.Println(err)
			return
		}

		_, err = database.MongoDB.Collection("token").InsertOne(c, models.Token{
			ID:                      primitive.NewObjectID(),
			Name:                    requestBody.Name,
			Count:                   count,
			Token:                   token,
			DateCreated:             primitive.NewDateTimeFromTime(time.Now()),
			DateExpires:             primitive.NewDateTimeFromTime(dateExpiresTime),
			CreatedBy:               c.MustGet("userIdPrimitive").(primitive.ObjectID),
			IsUserRegistrationToken: true,
			SuperUser:               requestBody.SuperUser,
			DirectAdd:               false,
			Used:                    0,
			Belongs:                 c.MustGet("userIdPrimitive").(primitive.ObjectID),
			UserGroup:               primitive.NilObjectID,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating token",
			})
			log.Println(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "Token created",
			"token":   token,
		})
	})

	router.POST("/manager/repoexists", middleware.LoginToken(), func(c *gin.Context) {
		//mark all as required
		var requestBody struct {
			GitRepo  string `json:"gitRepo" bson:"gitRepo" binding:"required"`
			GitOwner string `json:"gitOwner" bson:"gitOwner"`
			IsOwn    bool   `json:"isOwn" bson:"isOwn"`
		}

		err := c.BindJSON(&requestBody)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		usr := c.MustGet("user").(models.User)

		var owner = ""

		if requestBody.IsOwn {
			owner = usr.GitHubUsername
		} else {
			owner = requestBody.GitOwner
		}

		if git.CheckRepoExists(owner, usr.GitHubToken, requestBody.GitRepo) {
			c.JSON(200, gin.H{
				"message": "Repo exists",
			})
		} else {
			c.JSON(400, gin.H{
				"message": "Invalid git repo",
			})
		}
	})

	router.GET("/manager/isusercolabo/:id", middleware.LoginToken(), func(c *gin.Context) {
		userId := c.Param("id")

		userIdd, err := primitive.ObjectIDFromHex(userId)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			log.Println(err)
			return
		}

		//check if user exists
		var user models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, bson.M{
			"_id": userIdd,
		}).Decode(&user)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		usr := c.MustGet("user").(models.User)

		//get user group

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

		if git.CheckIfUserIsColabo(userGroup.GitHubOwner, user.GitHubUsername, usr.GitHubToken, userGroup.GitHubRepo) {
			c.JSON(200, gin.H{
				"message": "User is collaborator",
			})
		} else {
			c.JSON(400, gin.H{
				"message": "User is not collaborator",
			})
		}
	})

	///manager/git/" + {{.ID}} +"/addAll to add all users from a git repo
	router.GET("/manager/git/:id/addall", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		user := c.MustGet("user").(models.User)

		//find group
		addedBson := bson.M{
			"_id": idd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//get all users from repo group
		addedBson = bson.M{
			"userGroup": group.ID,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		users, err := database.MongoDB.Collection("gitUser").Find(c, addedBson)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching users",
			})
			return
		}

		//add all users to repo
		problems := ""

		for users.Next(c) {
			var gitUser models.GitHubUser
			err = users.Decode(&gitUser)
			if err != nil {
				log.Println(err)
				continue
			}

			if !gitUser.AddedToRepo {
				if !git.AddUserToRepo(gitUser.GitHubUsername, user.GitHubToken, group.GitHubRepo, group.GitHubOwner) {
					problems += gitUser.GitHubUsername + ", "
				}

				database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
					"_id": gitUser.ID,
				}, bson.M{
					"$set": bson.M{
						"addedToRepo": true,
					},
				})
			}
		}

		c.JSON(200, gin.H{
			"message":  "All users added to repo",
			"problems": problems,
		})
	})

	///manager/git/" + {{.ID}} +"/removeAll to remove all users from a git repo
	router.GET("/manager/git/:id/removeall", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		user := c.MustGet("user").(models.User)

		//find group
		addedBson := bson.M{
			"_id": idd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//check if the user is the owner of the repo
		if group.GitHubOwner != c.MustGet("user").(models.User).GitHubUsername {
			c.JSON(400, gin.H{
				"message": "You are not the owner of the repo",
			})
			return
		}

		//remove all users from repo
		addedBson = bson.M{
			"userGroup": group.ID,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		users, err := database.MongoDB.Collection("gitUser").Find(c, addedBson)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching users",
			})
			return
		}

		problems := ""

		for users.Next(c) {
			var gitUser models.GitHubUser
			err = users.Decode(&gitUser)
			if err != nil {
				log.Println(err)
				continue
			}

			if gitUser.AddedToRepo {
				if !git.RemoveUserFromRepo(group.GitHubOwner, gitUser.GitHubUsername, user.GitHubToken, group.GitHubRepo) {
					problems += gitUser.GitHubUsername + ", "
				}

				database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
					"_id": gitUser.ID,
				}, bson.M{
					"$set": bson.M{
						"addedToRepo": false,
					},
				})
			}
		}

		c.JSON(200, gin.H{
			"message":  "All users removed from repo",
			"problems": problems,
		})
	})

	///manager/git/" + {{.ID}} +"/remove/ {{.UserID}} to remove a user from a group
	router.GET("/manager/git/:id/remove/:userid", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		user := c.MustGet("user").(models.User)

		//find group
		addedBson := bson.M{
			"_id": idd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//find user
		userid := c.Param("userid")
		useridd, err := primitive.ObjectIDFromHex(userid)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		addedBson = bson.M{
			"_id":       useridd,
			"userGroup": group.ID,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var gitUser models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, addedBson).Decode(&gitUser)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		//remove user from repo
		if !git.RemoveUserFromRepo(group.GitHubOwner, gitUser.GitHubUsername, user.GitHubToken, group.GitHubRepo) {
			c.JSON(400, gin.H{
				"message": "The user could not be removed from the repo",
			})
			return
		}

		database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
			"_id": gitUser.ID,
		}, bson.M{
			"$set": bson.M{
				"addedToRepo": false,
			},
		}, options.Update())
		c.JSON(200, gin.H{
			"message": "User removed from repo",
		})
	})

	//manager/git/" + {{.ID}} +"/add/ {{.UserID}} to add a user to a repo - sents a invite first
	router.GET("/manager/git/:id/add/:userid", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		user := c.MustGet("user").(models.User)

		//find group
		addedBson := bson.M{
			"_id": idd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//find user
		userid := c.Param("userid")
		useridd, err := primitive.ObjectIDFromHex(userid)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		addedBson = bson.M{
			"_id":       useridd,
			"userGroup": group.ID,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var gitUser models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, addedBson).Decode(&gitUser)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		//add user to repo
		if !git.AddUserToRepo(gitUser.GitHubUsername, user.GitHubToken, group.GitHubRepo, group.GitHubOwner) {
			c.JSON(400, gin.H{
				"message": "The user could not be added to the repo",
			})
			return
		}

		database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
			"_id": gitUser.ID,
		}, bson.M{
			"$set": bson.M{
				"userGroup": group.ID,
			},
		}, options.Update())

		c.JSON(200, gin.H{
			"message": "User added to repo",
		})
	})

	///manager/group/" + {{.ID}} +"/removeall to remove all users from a group
	router.GET("/manager/group/:id/removeall", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		user := c.MustGet("user").(models.User)

		//find group
		addedBson := bson.M{
			"_id": idd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//remove all users from repo
		addedBson = bson.M{
			"userGroup": group.ID,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		users, err := database.MongoDB.Collection("gitUser").Find(c, addedBson)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching users",
			})
			return
		}

		problems := ""

		for users.Next(c) {
			var gitUser models.GitHubUser
			err = users.Decode(&gitUser)
			if err != nil {
				log.Println(err)
				continue
			}

			database.MongoDB.Collection("gitUser").UpdateOne(c,
				bson.M{
					"_id": gitUser.ID,
				},
				bson.M{
					"$set": bson.M{
						"userGroup": primitive.NilObjectID,
					},
				}, options.Update())
		}

		c.JSON(200, gin.H{
			"message":  "All users removed from repo",
			"problems": problems,
		})
	})

	///manager/group/" + {{.ID}} +"/remove/ {{.UserID}} to remove a user from a group
	router.GET("/manager/group/:id/remove/:userid", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		user := c.MustGet("user").(models.User)

		//find group
		addedBson := bson.M{
			"_id": idd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//find user
		userid := c.Param("userid")
		useridd, err := primitive.ObjectIDFromHex(userid)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		addedBson = bson.M{
			"_id":       useridd,
			"userGroup": group.ID,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var gitUser models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, addedBson).Decode(&gitUser)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
			"_id": gitUser.ID,
		}, bson.M{
			"$set": bson.M{
				"userGroup": primitive.NilObjectID,
			},
		}, options.Update())
	})

	//manager/group/" + {{.ID}} +"/add/ {{.UserID}} to add a user to a group
	router.GET("/manager/group/:id/add/:userid", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		user := c.MustGet("user").(models.User)

		//find group
		addedBson := bson.M{
			"_id": idd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, addedBson).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//find user
		userid := c.Param("userid")
		useridd, err := primitive.ObjectIDFromHex(userid)

		if err != nil {
			//wrong gitusr request
			c.JSON(400, gin.H{
				"message": "Invalid user id",
			})
			return
		}

		addedBson = bson.M{
			"_id": useridd,
		}

		if !user.IsSuperUser {
			addedBson["belongs"] = user.ID
		}

		var gitUser models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, addedBson).Decode(&gitUser)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		//check if user is already in a group
		isColabo := git.CheckIfUserIsColabo(group.GitHubOwner, gitUser.GitHubUsername, user.GitHubToken, group.GitHubRepo)

		database.MongoDB.Collection("gitUser").UpdateOne(c, bson.M{
			"_id": gitUser.ID,
		}, bson.M{
			"$set": bson.M{
				"userGroup":   group.ID,
				"addedToRepo": isColabo,
			},
		}, options.Update())

		c.JSON(200, gin.H{
			"message": "User added to group",
		})
	})

	router.GET("/profile", middleware.LoginToken(), func(c *gin.Context) {
		user := c.MustGet("user").(models.User)

		userData := userModalToData(user)

		template := template.Must(template.ParseFS(files, "main/public/profile/index.gohtml", "main/public/templates/template.gohtml"))
		template.Execute(c.Writer, userData)
	})

	//profile/update/email username git password

	router.POST("/profile/update/email", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Email string `json:"email" bson:"email" binding:"required"`
		}

		usr := c.MustGet("user").(models.User)

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		//check if email is valid
		err = joi.Validate(requestBody.Email, EmailSchema)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid email",
			})
			log.Println(err)
			return
		}

		//check if email is already in use
		var user models.User
		err = database.MongoDB.Collection("user").FindOne(c, bson.M{
			"email": requestBody.Email,
		}).Decode(&user)

		if err == nil {
			c.JSON(400, gin.H{
				"message": "Email is already in use",
			})
			log.Println("Email is already in use")
			return
		}

		//update email
		_, err = database.MongoDB.Collection("user").UpdateOne(c, bson.M{
			"_id": usr.ID,
		}, bson.M{
			"$set": bson.M{
				"email": requestBody.Email,
			},
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when updating email",
			})
			log.Println("Internal server error when updating email")
			return
		}

		c.JSON(200, gin.H{
			"message": "Email updated",
		})
	})

	router.POST("/profile/update/username", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Username string `json:"username" bson:"username" binding:"required"`
		}

		usr := c.MustGet("user").(models.User)

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		err = joi.Validate(requestBody.Username, UsernameSchema)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid username",
			})
			log.Println(err)
			return
		}

		//check if username is already in use
		var user models.User
		err = database.MongoDB.Collection("user").FindOne(c, bson.M{
			"username": requestBody.Username,
		}).Decode(&user)

		if err == nil {
			c.JSON(400, gin.H{
				"message": "Username is already in use",
			})
			log.Println("Username is already in use")
			return
		}

		//update username
		_, err = database.MongoDB.Collection("user").UpdateOne(c, bson.M{
			"_id": usr.ID,
		}, bson.M{
			"$set": bson.M{
				"username": requestBody.Username,
			},
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when updating username",
			})
			log.Println("Internal server error when updating username")
			return
		}

		c.JSON(200, gin.H{
			"message": "Username updated",
		})
	})

	//updates gitUsername and gitToken
	router.POST("/profile/update/git", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			GitUsername string `json:"gitUsername" bson:"gitUsername" binding:"required"`
			GitToken    string `json:"gitToken" bson:"gitToken" binding:"required"`
		}

		usr := c.MustGet("user").(models.User)

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		err = joi.Validate(requestBody.GitUsername, GitHubUsername)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitUsername",
			})
			log.Println(err)
			return
		}

		err = joi.Validate(requestBody.GitToken, GitTokenSchema)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitToken",
			})
			log.Println(err)
			return
		}

		if !git.CheckNewToken(requestBody.GitUsername, requestBody.GitToken, usr.GitHubToken) {
			c.JSON(400, gin.H{
				"message": "Invalid gitToken or gitUsername",
			})
			log.Println("Invalid gitToken or gitUsername")
			return
		}

		//update gitUsername and gitToken
		_, err = database.MongoDB.Collection("user").UpdateOne(c, bson.M{
			"_id": usr.ID,
		}, bson.M{
			"$set": bson.M{
				"githubUsername": requestBody.GitUsername,
				"githubToken":    requestBody.GitToken,
			},
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when updating gitUsername and gitToken",
			})
			log.Println("Internal server error when updating gitUsername and gitToken")
			return
		}

		c.JSON(200, gin.H{
			"message": "GitUsername and gitToken updated",
		})
	})

	router.POST("/profile/update/password", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Password string `json:"password" bson:"password" binding:"required"`
		}

		usr := c.MustGet("user").(models.User)

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			log.Println(err)
			return
		}

		err = joi.Validate(requestBody.Password, PasswordSchema)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid password",
			})
			log.Println(err)
			return
		}

		//hash password
		hashedPassword, err := crypt.HashPassword(requestBody.Password)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when hashing password",
			})
			log.Println("Internal server error when hashing password")
			return
		}

		//update password
		_, err = database.MongoDB.Collection("user").UpdateOne(c, bson.M{
			"_id": usr.ID,
		}, bson.M{
			"$set": bson.M{
				"password": hashedPassword,
			},
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when updating password",
			})
			log.Println("Internal server error when updating password")
			return
		}

		c.JSON(200, gin.H{
			"message": "Password updated",
		})
	})

	router.DELETE("/manager/notification/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")

		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid notification id",
			})
			log.Println(err)
			return
		}

		_, err = database.MongoDB.Collection("notification").DeleteOne(c, bson.M{
			"_id": idd,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when deleting notification",
			})
			log.Println(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "Notification deleted",
		})
	})

	router.GET("/manager/group/:id/scan", middleware.LoginToken(), func(c *gin.Context) {
		usr := c.MustGet("user").(models.User)

		//get group from id
		id := c.Param("id")
		idd, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid group id",
			})
			return
		}

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id": idd,
		}).Decode(&group)

		var groupsWithRepo []models.UserGroup
		cur, err := database.MongoDB.Collection("userGroup").Find(c, bson.M{
			"gitHubOwner": group.GitHubOwner,
			"gitHubRepo":  group.GitHubRepo,
		})

		for cur.Next(c) {
			var group models.UserGroup
			err = cur.Decode(&group)
			if err != nil {
				log.Println(err)
				continue
			}

			groupsWithRepo = append(groupsWithRepo, group)
		}

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Group not found",
			})
			return
		}

		//check if the user is the owner of the repo
		if group.GitHubOwner != usr.GitHubUsername {
			c.JSON(400, gin.H{
				"message": "You are not the owner of the repo",
			})
			return
		}

		//get all members of repo
		members := git.GetColabosFromRepo(usr.GitHubToken, group.GitHubOwner, group.GitHubRepo)

		membersInDB := []models.GitHubUser{}
		membersInGrpDB := []models.GitHubUser{}

		//get all members in db
		cur, err = database.MongoDB.Collection("gitUser").Find(c, bson.M{
			"belongs": usr.ID,
		})

		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching members",
			})
		}

		for cur.Next(c) {
			var member models.GitHubUser
			err = cur.Decode(&member)
			if err != nil {
				log.Println(err)
				continue
			}

			inGrp := false
			for _, group := range groupsWithRepo {
				if member.UserGroup == group.ID {
					inGrp = true
					break
				}
			}

			if !inGrp {
				continue
			}

			membersInDB = append(membersInDB, member)

			if member.UserGroup == group.ID {
				membersInGrpDB = append(membersInGrpDB, member)
			}
		}

		//remove duplicated members from membersInDB
		for i, member := range membersInDB {
			for _, memberInGrp := range membersInGrpDB {
				if member.GitHubUsername == memberInGrp.GitHubUsername {
					membersInDB = append(membersInDB[:i], membersInDB[i+1:]...)
					break
				}
			}
		}

		type Member struct {
			Name    string
			GitName string
			Url     string
			Email   string
			Icon    string
		}

		type Data struct {
			Members        string
			MembersInGroup string
			Repo           string
			Owner          string
			GroupName      string
			GroupID        string
		}

		membersInGrpParse := []Member{}
		membersInDbParse := []Member{}

		for _, member := range members {
			if member.GetLogin() == group.GitHubOwner {
				continue
			}

			existsDb := false
			for _, memberInDB := range membersInDB {
				if memberInDB.GitHubUsername == member.GetLogin() {
					existsDb = true
					continue
				}
			}

			existsGrp := false
			for _, memberInGrpDB := range membersInGrpDB {
				if memberInGrpDB.GitHubUsername == member.GetLogin() {
					existsGrp = true
					continue
				}
			}

			if !existsGrp {
				membersInGrpParse = append(membersInGrpParse, Member{
					Name:    member.GetName(),
					GitName: member.GetLogin(),
					Url:     member.GetHTMLURL(),
					Email:   member.GetEmail(),
					Icon:    member.GetAvatarURL(),
				})
			}

			if !existsDb {
				membersInDbParse = append(membersInDbParse, Member{
					Name:    member.GetName(),
					GitName: member.GetLogin(),
					Url:     member.GetHTMLURL(),
					Email:   member.GetEmail(),
					Icon:    member.GetAvatarURL(),
				})
			}
		}

		jsonData, _ := json.Marshal(membersInDbParse)
		jsonData2, _ := json.Marshal(membersInGrpParse)

		data := Data{
			Members:        string(jsonData),
			Repo:           group.GitHubRepo,
			Owner:          group.GitHubOwner,
			GroupName:      group.Name,
			GroupID:        group.ID.Hex(),
			MembersInGroup: string(jsonData2),
		}

		template := template.Must(template.ParseFS(files, "main/public/manager/group/scan/index.gohtml", "main/public/templates/template.gohtml"))
		template.Execute(c.Writer, data)
	})
}
