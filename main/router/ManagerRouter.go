package router

import (
	"ASO/main/database"
	"ASO/main/database/models"
	"ASO/main/git"
	"ASO/main/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"text/template"
	"time"
)

func fetchAllGitUsers(c *gin.Context) ([]GitUserData, error) {
	cur, err := database.MongoDB.Collection("gitUser").Find(c, bson.M{
		"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
	})
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
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id":     user.UserGroup,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&userGroup)

		var usrDt GitUserData

		if err != nil {
			usrDt = GitUserData{
				ID:             user.ID.Hex(),
				GitHubUsername: user.GitHubUsername,
				DateCreated:    user.DateCreated.Time().Format("2006-01-02 15:04:05"),
				DateExpires:    user.DateExpires.Time().Format("2006-01-02 15:04:05"),
				Username:       user.Username,
			}
		} else {
			usrDt = GitUserData{
				ID:             user.ID.Hex(),
				GitHubUsername: user.GitHubUsername,
				DateCreated:    user.DateCreated.Time().Format("2006-01-02 15:04:05"),
				DateExpires:    user.DateExpires.Time().Format("2006-01-02 15:04:05"),
				UserGroup:      userGroupModalToData(userGroup),
				Username:       user.Username,
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
	var g = UserGroupData{
		ID:          group.ID.Hex(),
		Name:        group.Name,
		Date:        group.Date.Time().Format("2006-01-02 15:04:05"),
		DateExpires: group.DateExpires.Time().Format("2006-01-02 15:04:05"),
		Expires:     group.Expires,
		AutoDelete:  group.AutoDelete,
		Notify:      group.Notify,
	}

	if g.DateExpires == "0001-01-01 01:00:00" {
		g.DateExpires = "Never"
	}
	return g
}

func tokenModalToData(token models.Token) TokenData {
	return TokenData{
		ID:          token.ID.Hex(),
		Name:        token.Name,
		Count:       token.Count,
		Token:       token.Token,
		DateCreated: token.DateCreated.Time().Format("2006-01-02 15:04:05"),
		DateExpires: token.DateExpires.Time().Format("2006-01-02 15:04:05"),
		DirectAdd:   token.DirectAdd,
		Used:        token.Used,
	}
}

func fetchAllGroups(c *gin.Context) ([]UserGroupData, error) {
	cur, err := database.MongoDB.Collection("userGroup").Find(c, bson.M{
		"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
	})
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
	cur, err := database.MongoDB.Collection("token").Find(c, bson.M{
		"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
	})
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
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id":     token.UserGroup,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&userGroup)

		grpErr := false
		if err != nil {
			grpErr = true
		}

		//get createdBy
		var createdBy models.User
		err = database.MongoDB.Collection("gitusr").FindOne(c, bson.M{
			"_id":     token.CreatedBy,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&createdBy)

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

type ManagerRouteData struct {
	Users  []GitUserData
	Groups []UserGroupData
	Tokens []TokenData
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
}

type UserGroupData struct {
	ID          string        `json:"id" bson:"_id"`
	Name        string        `json:"name" bson:"name"`
	Date        string        `json:"date" bson:"date"`
	DateExpires string        `json:"dateExpires" bson:"dateExpires"`
	Members     []GitUserData `json:"members" bson:"members"`
	Users       int           `json:"users" bson:"users"`
	Expires     bool          `json:"expires" bson:"expires"`
	AutoDelete  bool          `json:"autoDelete" bson:"autoDelete"`
	Notify      bool          `json:"notify" bson:"notify"`
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
				if users[usr].UserGroup.ID == groups[grp].ID {
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

		data := ManagerRouteData{
			Groups: groups,
			Users:  users,
			Tokens: tokens,
		}

		template := template.Must(template.ParseFiles("main/public/manager/index.gohtml"))
		template.Execute(c.Writer, data)
	})

	router.GET("/manager/gitusr/:id", middleware.LoginToken(), func(c *gin.Context) {
		id := c.Param("id")

		//check if id is valid
		idd, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitusr id",
			})
			return
		}

		var user models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, bson.M{
			"_id":     idd,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&user)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User not found",
			})
			return
		}

		//fetch gitusr group from gitusr
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id":     user.UserGroup,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&userGroup)

		//fetch all groups
		grps, err := fetchAllGroups(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching groups",
			})
			return
		}

		var userData GitUserData

		if err != nil {
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
			userData = GitUserData{
				ID:             user.ID.Hex(),
				GitHubUsername: user.GitHubUsername,
				Username:       user.Username,
				ExpiryByGroup:  user.ExpiresGroup,
				DateCreated:    user.DateCreated.Time().Format("2006-01-02 15:04:05"),
				DateExpires:    user.DateExpires.Time().Format("2006-01-02 15:04:05"),
				UserGroup:      userGroupModalToData(userGroup),
				Groups:         grps,
			}
		}

		if userData.DateExpires == "0001-01-01 01:00:00" {
			userData.DateExpires = "Never"
		}

		fmt.Println(userData)

		template := template.Must(template.ParseFiles("main/public/manager/gitusr/index.gohtml"))
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

		var group models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id":     idd,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&group)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User group not found",
			})
			fmt.Println(err)
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
				members = append(members, users[i])
			}
		}

		groupData := userGroupModalToData(group)

		if len(members) > 0 {
			groupData.Members = members
		}

		fmt.Println(groupData)

		template := template.Must(template.ParseFiles("main/public/manager/group/index.gohtml"))
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
		template := template.Must(template.ParseFiles("main/public/manager/group/create/index.gohtml"))
		template.Execute(c.Writer, nil)
	})

	router.POST("/manager/group/create", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Name        string `json:"name" bson:"name"`
			DateExpires string `json:"dateExpires" bson:"dateExpires"`
			Notify      bool   `json:"notify" bson:"notify"`
			Expires     bool   `json:"doesExpire" bson:"doesExpire"`
			AutoDelete  bool   `json:"autoDelete" bson:"autoDelete"`
		}
		//get from body
		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			fmt.Println(err)
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04", requestBody.DateExpires)

		if requestBody.Expires {
			if err != nil || dateExpiresTime.Before(time.Now()) {
				c.JSON(400, gin.H{
					"message": "Invalid date",
				})
				fmt.Println(err)
				return
			}
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
			Expires:         requestBody.Expires,
			Belongs:         c.MustGet("userIdPrimitive").(primitive.ObjectID),
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating group",
			})
			fmt.Println(err)
			return
		}

		c.JSON(200, gin.H{
			"message": "Group created",
		})
	})

	router.GET("/manager/token/create", middleware.LoginToken(), func(c *gin.Context) {
		grps, err := fetchAllGroups(c)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when fetching groups",
			})
			return
		}

		template := template.Must(template.ParseFiles("main/public/manager/token/create/index.gohtml"))
		template.Execute(c.Writer, ManagerCreateTkData{
			Groups: grps,
		})
	})

	router.POST("/manager/token/create", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Name        string `json:"name" bson:"name"`
			UserGroup   string `json:"userGroup" bson:"userGroup"`
			DirectAdd   bool   `json:"directAdd" bson:"directAdd"`
			AutoDelete  bool   `json:"autoDelete" bson:"autoDelete"`
			Notify      bool   `json:"notify" bson:"notify"`
			DateExpires string `json:"dateExpires" bson:"dateExpires"`
			Count       string `json:"count" bson:"count"`
		}

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			fmt.Println(err)
			return
		}

		//check UserGroup
		grp, err := primitive.ObjectIDFromHex(requestBody.UserGroup)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitusr group id",
			})
			fmt.Println(err)
			return
		}

		//check if exists
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id":     grp,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&userGroup)

		//parse count
		count, err := strconv.Atoi(requestBody.Count)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid count",
			})
			fmt.Println(err)
			return
		}
		//check count
		if count < 1 {
			c.JSON(400, gin.H{
				"message": "Invalid count",
			})

			fmt.Println("Invalid count")
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04:05", requestBody.DateExpires)

		if err != nil || dateExpiresTime.Before(time.Now()) {
			c.JSON(400, gin.H{
				"message": "Invalid date",
			})
			fmt.Println(err)
			return
		}

		token := GenerateRandomString(6)

		//create token
		_, err = database.MongoDB.Collection("token").InsertOne(c, models.Token{
			ID:          primitive.NewObjectID(),
			Name:        requestBody.Name,
			Count:       count,
			Token:       token,
			UserGroup:   userGroup.ID,
			DateCreated: primitive.NewDateTimeFromTime(time.Now()),
			DateExpires: primitive.NewDateTimeFromTime(dateExpiresTime),
			CreatedBy:   c.MustGet("userIdPrimitive").(primitive.ObjectID),
			IsReg:       false,
			DirectAdd:   requestBody.DirectAdd,
			Used:        0,
			Belongs:     c.MustGet("userIdPrimitive").(primitive.ObjectID),
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating token",
			})
			fmt.Println(err)
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
		var tk models.Token
		err := database.MongoDB.Collection("token").FindOne(c, bson.M{
			"token":   token,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&tk)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Token not found",
			})
			return
		}

		template := template.Must(template.ParseFiles("main/public/token/index.gohtml"))
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

		var token models.Token
		err = database.MongoDB.Collection("token").FindOne(c, bson.M{
			"_id":     idd,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&token)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Token not found",
			})
			return
		}

		tk := tokenModalToData(token)

		template := template.Must(template.ParseFiles("main/public/manager/token/index.gohtml"))
		template.Execute(c.Writer, tk)
	})

	router.GET("/tk/:tk", func(c *gin.Context) {
		token := c.Param("tk")

		//check if token exists

		var tk models.Token
		err := database.MongoDB.Collection("token").FindOne(c, bson.M{
			"token":   token,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&tk)

		failed := false
		message := "no message"

		if err != nil {
			failed = true
			message = "Token not found"
		}

		//check if token is expired
		if tk.DateExpires.Time().Before(time.Now()) {
			failed = true
			message = "Token is already expired"
		}

		//check if token is used
		if tk.Used >= tk.Count {
			failed = true
			message = "Tokens usage limit reached"
		}

		template := template.Must(template.ParseFiles("main/public/tk/index.gohtml"))
		template.Execute(c.Writer, ManagerTkData{
			Failed:  failed,
			Message: message,
		})
	})

	router.POST("/tk", func(c *gin.Context) {
		var requestBody struct {
			Token          string `json:"token" bson:"token"`
			GitHubUsername string `json:"gitUsername" bson:"gitUsername"`
			Email          string `json:"email" bson:"email"`
			Username       string `json:"username" bson:"username"`
		}

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			fmt.Println(err)
			return
		}

		//check if token exists

		var tk models.Token
		err = database.MongoDB.Collection("token").FindOne(c, bson.M{
			"token":   requestBody.Token,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&tk)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "Token not found",
			})
			return
		}

		//check if token is expired
		if tk.DateExpires.Time().Before(time.Now()) {
			c.JSON(400, gin.H{
				"message": "Token is already expired",
			})
			fmt.Println("Token is already expired")
			return
		}

		//check if token is used
		if tk.Used >= tk.Count {
			c.JSON(400, gin.H{
				"message": "Tokens usage limit reached",
			})
			fmt.Println("Tokens usage limit reached")
			return
		}

		fmt.Println("checking")

		var user models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, bson.M{
			"githubUsername": requestBody.GitHubUsername,
			"belongs":        c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&user)

		if err == nil {
			c.JSON(400, gin.H{
				"message": "User already exists",
			})
			fmt.Println("User already exists")
			return
		}

		//check if gitusr exists by email
		err = database.MongoDB.Collection("gitUser").FindOne(c, bson.M{
			"email":   requestBody.Email,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&user)

		if err == nil {
			c.JSON(400, gin.H{
				"message": "User already exists",
			})
			fmt.Println("User already exists")
			return
		}

		//get userGroup
		var userGroup models.UserGroup
		err = database.MongoDB.Collection("userGroup").FindOne(c, bson.M{
			"_id":     tk.UserGroup,
			"belongs": c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&userGroup)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "User group not found",
			})
			fmt.Println("User group not found")
			return
		}

		tbu := git.CheckUser(requestBody.Email, requestBody.GitHubUsername)
		if tbu == "" {
			c.JSON(400, gin.H{
				"message": "User can not be found on github",
			})
			fmt.Println("User already exists")
			return
		}

		if tk.DirectAdd {
			if !git.AddUserToRepo(tbu) {
				c.JSON(500, gin.H{
					"message": "Internal server error when adding gitusr to repo",
				})
				fmt.Println("Internal server error when adding gitusr to repo")
				return
			}
		}

		//check if the username is okay length, characters etc
		username := requestBody.Username

		if len(username) < 3 {
			c.JSON(400, gin.H{
				"message": "Username is too short",
			})
			fmt.Println("Username is too short")
			return
		}

		if len(username) > 20 {
			c.JSON(400, gin.H{
				"message": "Username is too long",
			})
			fmt.Println("Username is too long")
			return
		}

		//create gitusr
		_, err = database.MongoDB.Collection("gitUser").InsertOne(c, models.GitHubUser{
			ID:             primitive.NewObjectID(),
			Username:       username,
			GitHubUsername: tbu,
			DateCreated:    primitive.NewDateTimeFromTime(time.Now()),
			Expires:        true,
			ExpiresGroup:   true,
			DateExpires:    userGroup.DateExpires,
			UserGroup:      userGroup.ID,
			AddedToRepo:    tk.DirectAdd,
			Belongs:        c.MustGet("userIdPrimitive").(primitive.ObjectID),
		})

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error when creating gitusr",
			})
			fmt.Println("Internal server error when creating gitusr")
			return
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

		template := template.Must(template.ParseFiles("main/public/manager/gitusr/create/index.gohtml"))
		template.Execute(c.Writer, ManagerCreateTkData{
			Groups: grps,
		})
	})

	router.POST("/manager/gitusr/create", middleware.LoginToken(), func(c *gin.Context) {
		var requestBody struct {
			Username    string `json:"username" bson:"username"`
			GitUsername string `json:"gitUsername" bson:"gitUsername"`
			UserGroup   string `json:"userGroup" bson:"userGroup"`
			ExpireGroup bool   `json:"expireGroup" bson:"expireGroup"`
			Expires     bool   `json:"expires" bson:"expires"`
			DateExpires string `json:"dateExpires" bson:"dateExpires"`
		}

		err := c.BindJSON(&requestBody)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid form data",
			})
			fmt.Println(err)
			return
		}

		//check if gitusr exists
		var user models.GitHubUser
		err = database.MongoDB.Collection("gitUser").FindOne(c, bson.M{
			"githubUsername": requestBody.GitUsername,
			"belongs":        c.MustGet("userIdPrimitive").(primitive.ObjectID),
		}).Decode(&user)

		if err == nil {
			c.JSON(400, gin.H{
				"message": "User already exists",
			})
			fmt.Println("User already exists")
			return
		}

		//check if gitusr exists by email skip

		//check if gitusr group exists
		grp, err := primitive.ObjectIDFromHex(requestBody.UserGroup)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid gitusr group id",
			})
			fmt.Println(err)
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
			fmt.Println("User group not found")
			return
		}

		//check if the username is okay length, characters etc
		username := requestBody.Username
		if len(username) < 3 {
			c.JSON(400, gin.H{
				"message": "Username is too short",
			})
			fmt.Println("Username is too short")
			return
		}

		if len(username) > 20 {
			c.JSON(400, gin.H{
				"message": "Username is too long",
			})
			fmt.Println("Username is too long")
			return
		}

		//parse expire date
		dateExpiresTime, err := time.Parse("2006-01-02T15:04:05", requestBody.DateExpires)

		if requestBody.Expires && !requestBody.ExpireGroup {
			if err != nil || dateExpiresTime.Before(time.Now()) {
				c.JSON(400, gin.H{
					"message": "Invalid date",
				})
				fmt.Println(err)
				return
			}
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
			fmt.Println("Internal server error when creating gitusr")
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

	///manager/group/" + {{.ID}} +"/removeAll to remove all users from a group
	///manager/group/" + {{.ID}} +"/remove/ {{.UserID}} to remove a user from a group
	//manager/group/" + {{.ID}} +"/add/ {{.UserID}} to add a user to a group
}
