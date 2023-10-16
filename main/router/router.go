package router

import (
	"ASO/main/crypt"
	"ASO/main/database"
	"ASO/main/database/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"net/http"
	"text/template"
	"time"
)

var register = false

func GenerateRandomString(length int) string {
	// Define the character set for lowercase and uppercase letters
	allLetters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		// Generate a random index within the range of the character set
		index := rand.Intn(len(allLetters))

		// Assign the character at the random index to the result
		result[i] = allLetters[index]
	}

	return string(result)
}

func InitRouter() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		if !register {
			//check if the regauth coockie exists
			regAuth, err := c.Cookie("regauth")

			cursor, err := database.MongoDB.Collection("user").Find(c, bson.M{}, options.Find())
			if err != nil {
				// Handle the error
				log.Fatal(err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
				return
			}
			defer cursor.Close(c)

			var users []models.User

			for cursor.Next(c) {
				var user models.User
				if err := cursor.Decode(&user); err != nil {
					log.Fatal(err)
				}
				users = append(users, user)
			}

			if err := cursor.Err(); err != nil {
				log.Fatal(err)
			}

			if len(users) == 0 {
				if regAuth != "" && err == nil {
					//redirect to the reg page
					c.Redirect(http.StatusTemporaryRedirect, "/reg")
					return
				}

				// No users found in the database, give permission
				token := GenerateRandomString(30)
				jwt, err := crypt.GenerateRegToken(token)
				if err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
					return
				}
				c.SetCookie("regauth", jwt, 3600*5, "/", "", false, false)
				dbTk := models.Token{
					ID:          primitive.NewObjectID(),
					Token:       token,
					IsReg:       true,
					DateCreated: primitive.NewDateTimeFromTime(time.Now()),
					DateExpires: primitive.NewDateTimeFromTime(time.Now().Add(time.Minute * 5)),
				}

				database.MongoDB.Collection("token").InsertOne(c, dbTk, options.InsertOne())

				c.Redirect(http.StatusTemporaryRedirect, "/reg")
				return
			} else {
				register = true
			}
		}
		if register {
			//check for auth cookie
			auth, err := c.Cookie("auth")
			if auth != "" && err == nil {
				//redirect to the reg page
				c.Redirect(http.StatusTemporaryRedirect, "/manager")
				return
			} else {
				template := template.Must(template.ParseFiles("main/public/homepage/index.gohtml"))
				template.Execute(c.Writer, nil)
			}
		}
	})

	router.GET("/reg", func(c *gin.Context) {
		cursor, err := database.MongoDB.Collection("user").Find(c, bson.M{}, options.Find())
		if err != nil {
			// Handle the error
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}
		defer cursor.Close(c)

		var users []models.User

		for cursor.Next(c) {
			var user models.User
			if err := cursor.Decode(&user); err != nil {
				log.Fatal(err)
			}
			users = append(users, user)
		}

		if err := cursor.Err(); err != nil {
			log.Fatal(err)
		}

		if len(users) == 0 {
			regAuth, err := c.Cookie("regauth")
			if regAuth != "" && err == nil {
				jwt, err := crypt.ParseJwt(regAuth)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
					fmt.Println(err)
					return
				}
				token := jwt["token"]

				res := database.MongoDB.Collection("token").FindOne(c, bson.M{
					"token": token,
				})

				if res.Err() != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
					return
				}

				var tk models.Token
				res.Decode(&tk)

				//check date
				if tk.DateExpires.Time().Before(time.Now()) {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
					return
				}
			}

			template := template.Must(template.ParseFiles("main/public/reg/index.gohtml"))
			template.Execute(c.Writer, nil)
		} else {
			c.Redirect(http.StatusTemporaryRedirect, "/")
		}
	})

	router.POST("/reg", func(c *gin.Context) {
		cursor, err := database.MongoDB.Collection("user").Find(c, bson.M{}, options.Find())
		if err != nil {
			// Handle the error
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}
		defer cursor.Close(c)

		var users []models.User

		for cursor.Next(c) {
			var user models.User
			if err := cursor.Decode(&user); err != nil {
				log.Fatal(err)
			}
			users = append(users, user)
		}

		if err := cursor.Err(); err != nil {
			log.Fatal(err)
		}

		if len(users) == 0 {
			regAuth, err := c.Cookie("regauth")
			var token = ""
			if regAuth != "" && err == nil {
				jwt, err := crypt.ParseJwt(regAuth)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
					fmt.Println(err)
					return
				}
				token = jwt["token"].(string)

				res := database.MongoDB.Collection("token").FindOne(c, bson.M{
					"token": token,
				})

				if res.Err() != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
					return
				}

				var tk models.Token
				res.Decode(&tk)

				//check date
				if tk.DateExpires.Time().Before(time.Now()) {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
					return
				}
			}

			var requestBody struct {
				Username string `json:"username" binding:"required"`
				Password string `json:"password" binding:"required"`
			}

			if err := c.ShouldBindJSON(&requestBody); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			username := requestBody.Username
			password := requestBody.Password

			// Check if the user already exists in the database by querying with the username
			var existingUser models.User
			filter := bson.M{"username": username}
			err = database.MongoDB.Collection("user").FindOne(c, filter).Decode(&existingUser)

			if err == nil {
				// User with the same username already exists
				c.JSON(http.StatusConflict, gin.H{"message": "User already exists"})
				return
			} else if err != mongo.ErrNoDocuments {
				// Handle other database query errors
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
				return
			}

			hashedPassword, err := crypt.HashPassword(password)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
				return
			}

			newUser := models.User{
				ID:       primitive.NewObjectID(),
				Username: username,
				Password: hashedPassword,
			}

			_, err = database.MongoDB.Collection("user").InsertOne(c, newUser, options.InsertOne())

			if err != nil {
				// Handle database insertion error
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
				return
			}

			_, _ = database.MongoDB.Collection("token").DeleteOne(c, bson.M{
				"token": token,
			})

			c.SetCookie("regauth", "", -1, "/", "", false, true)

			c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Created user"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
		}
	})

	router.GET("/login", func(c *gin.Context) {
		//check auth cookie, if exists redirect to manager
		auth, err := c.Cookie("auth")
		if auth != "" && err == nil {
			//redirect to the reg page
			c.Redirect(http.StatusTemporaryRedirect, "/manager")
			return
		} else {
			template := template.Must(template.ParseFiles("main/public/login/index.gohtml"))
			template.Execute(c.Writer, nil)
		}
	})

	router.POST("/login", func(c *gin.Context) {
		//check body for username and password
		var requestBody struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		username := requestBody.Username
		password := requestBody.Password

		//check if user exists
		var user models.User
		filter := bson.M{"username": username}
		err := database.MongoDB.Collection("user").FindOne(c, filter).Decode(&user)

		//if user exists, check password using crypt
		if err == nil {
			if crypt.CheckPasswordHash(password, user.Password) {
				//generate jwt token
				token, err := crypt.GenerateLoginToken(user.ID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
					return
				}

				//set cookie with age of 2 days, setting maxAge to: 3600 * 24 * 2
				c.SetCookie("auth", token, 3600*24*2, "/", "", false, false)

				c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Logged in"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			}
		}
	})

	// Seite zur Generierung des Tokens
	router.GET("/generate-token", func(c *gin.Context) {
		// Hier können Sie die Logik zur Generierung des GitHub-Tokens implementieren
		// und die generierten Tokens an die Benutzer senden.
		// Zum Beispiel: c.JSON(http.StatusOK, gin.H{"token": "GENERATED_TOKEN"})
	})

	// Seite zur Eingabe des Tokens und Hinzufügen des Benutzernamens zum Repo
	router.GET("/token", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(400, gin.H{
				"message": "Please provide a token",
			})
			return
		} else {

		}
	})

	initManagerRouter(router)

	router.Run(":8080")
}
