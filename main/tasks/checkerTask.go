package tasks

import (
	"ASO/main/database"
	"ASO/main/database/models"
	"ASO/main/git"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func initCheckerTasks() {
	_, err := Scheduler.Every(1).Hours().Do(func() {
		checkTokens()
		checkGitUsers()
		checkGroups()
	})

	if err != nil {
		fmt.Println("Failed to start checker task")
		fmt.Println(err)
	}
}

func checkTokens() {
	fmt.Println("Checking tokens")
	cur, err := database.MongoDB.Collection("token").Find(context.Background(), bson.M{
		"$or": []bson.M{
			{
				"$expr": bson.M{
					"$gte": []interface{}{"$used", "$count"},
				},
			},
			{
				"dateExpires": bson.M{
					"$lte": time.Now(), // Check if "dateExpires" is less than or equal to the current time
				},
			},
		},
	})

	if err != nil {
		fmt.Println("Failed to get tokens")
		fmt.Println(err)
	}

	var tokens []models.Token
	err = cur.All(context.Background(), &tokens)

	if err != nil {
		fmt.Println("!!Failed to parse tokens!!")
		fmt.Println(err)
	}

	for _, token := range tokens {
		fmt.Println("Token expired: " + token.Name)

		//get creator
		var creator models.User

		if token.CreatedBy != primitive.NilObjectID {
			err = database.MongoDB.Collection("user").FindOne(context.Background(), bson.M{
				"_id": token.Belongs,
			}).Decode(&creator)

			if err != nil {
				fmt.Println("Failed to get creator")
				fmt.Println(err)
				continue
			}
		}

		_, err = database.MongoDB.Collection("token").DeleteOne(context.Background(), bson.M{
			"_id": token.ID,
		})

		if err != nil {
			fmt.Println("Failed to delete token " + token.Name)
			fmt.Println(err)
			continue
		}

		if token.CreatedBy != primitive.NilObjectID {
			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      token.Belongs,
				Notification: "Token " + token.Name + " has been deleted!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Token deleted",
				UserGroup:    primitive.NilObjectID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "warning",
			})
		}
	}
}

func checkGitUsers() {
	fmt.Println("Checking git users")
	cur, err := database.MongoDB.Collection("githubUser").Find(context.Background(), bson.M{
		"expires":      true,
		"expiresGroup": false,
		"dateExpires": bson.M{
			"$lte": time.Now(),
		},
	})

	if err != nil {
		fmt.Println("Failed to get users")
		fmt.Println(err)
	}

	var users []models.GitHubUser
	err = cur.All(context.Background(), &users)

	if err != nil {
		fmt.Println("!!Failed to parse users!!")
		fmt.Println(err)
	}

	for _, user := range users {
		fmt.Println("User expired: " + user.Username)

		//get creator
		var creator models.User

		err = database.MongoDB.Collection("user").FindOne(context.Background(), bson.M{
			"_id": user.Belongs,
		}).Decode(&creator)

		if err != nil {
			fmt.Println("Failed to get creator")
			fmt.Println(err)
			continue
		}

		//get group
		var group models.UserGroup

		err = database.MongoDB.Collection("userGroup").FindOne(context.Background(), bson.M{
			"_id": user.UserGroup,
		}).Decode(&group)

		if err == nil {
			if user.AddedToRepo {
				if !git.RemoveUserFromRepo(creator.GitHubUsername, user.GitHubUsername, creator.GitHubToken, user.UserGroup.Hex()) {
					fmt.Println("Failed to remove user " + user.Username + " from repo " + user.UserGroup.Hex())
					sendErrorNotification("Failed to remove user from repo", "Failed to remove user "+user.Username+" from repo "+group.GitHubRepo, user.Belongs, 1, user.ID)
					continue
				}
			}
		}

		_, err = database.MongoDB.Collection("githubUser").DeleteOne(context.Background(), bson.M{
			"_id": user.ID,
		})

		if err != nil {
			fmt.Println("Failed to delete user " + user.Username)
			fmt.Println(err)
			continue
		}

		database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
			ID:           primitive.NewObjectID(),
			Belongs:      user.Belongs,
			Notification: "User " + user.Username + " has been deleted!",
			DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
			Title:        "User deleted",
			UserGroup:    primitive.NilObjectID,
			GitHubUser:   primitive.NilObjectID,
			Token:        primitive.NilObjectID,
			Style:        "warning",
		})
	}
}

func checkGroups() {
	fmt.Println("Checking groups")
	cur, err := database.MongoDB.Collection("userGroup").Find(context.Background(), bson.M{
		"expires":  true,
		"notified": false,
		"dateExpires": bson.M{
			"$lte": time.Now(),
		},
	})

	if err != nil {
		fmt.Println("Failed to get groups")
		fmt.Println(err)
	}

	var groups []models.UserGroup
	err = cur.All(context.Background(), &groups)

	if err != nil {
		fmt.Println("!!Failed to parse groups!!")
		fmt.Println(err)
	}

	for _, group := range groups {
		fmt.Println("Group expired: " + group.Name)

		//get users in group
		cur, err := database.MongoDB.Collection("gitUser").Find(context.Background(), bson.M{
			"userGroup": group.ID,
		})

		if err != nil {
			fmt.Println("Failed to get users")
			fmt.Println(err)
			continue
		}

		var users []models.GitHubUser
		err = cur.All(context.Background(), &users)

		if err != nil {
			fmt.Println("Failed to parse users")
			fmt.Println(err)
			continue
		}

		//get creator
		var creator models.User

		err = database.MongoDB.Collection("user").FindOne(context.Background(), bson.M{
			"_id": group.Belongs,
		}).Decode(&creator)

		if err != nil {
			fmt.Println("Failed to get creator")
			fmt.Println(err)
			continue
		}

		if group.AutoRemoveUsers || group.AutoDelete {
			for _, user := range users {
				if user.ExpiresGroup && group.AutoRemoveUsers {
					failed := false
					if user.AddedToRepo {
						if !git.RemoveUserFromRepo(group.GitHubOwner, user.GitHubUsername, creator.GitHubToken, group.GitHubRepo) {
							failed = true
						}
					}

					if failed {
						fmt.Println("Failed to remove user " + user.Username + " from repo " + group.GitHubRepo)
						sendErrorNotification("Failed to remove user from repo", "Failed to remove user "+user.Username+" from repo "+group.GitHubRepo, user.Belongs, 1, user.ID)
						continue
					}

					_, err = database.MongoDB.Collection("gitUser").DeleteOne(context.Background(), bson.M{
						"_id": user.ID,
					})

					if err != nil {
						fmt.Println("Failed to delete user " + user.Username)
						fmt.Println(err)
						continue
					}
				} else {
					_, err = database.MongoDB.Collection("gitUser").UpdateOne(context.Background(), bson.M{
						"_id": user.ID,
					}, bson.M{
						"$set": bson.M{
							"userGroup": primitive.NilObjectID,
						},
					})

					if err != nil {
						fmt.Println("Failed to update user " + user.Username)
						fmt.Println(err)
						continue
					}
				}
			}
		}

		if group.AutoDelete {
			repoRemoveText := ""
			if group.AutoRemoveUsers {
				repoRemoveText = " and removed from repo " + group.GitHubRepo
			}

			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      group.Belongs,
				Notification: "Group " + group.Name + " has been deleted! " + fmt.Sprintf("%d", len(users)) + " users have been removed from the group" + repoRemoveText + "!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Group deleted",
				UserGroup:    primitive.NilObjectID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "warning",
			})

			_, err = database.MongoDB.Collection("userGroup").DeleteOne(context.Background(), bson.M{
				"_id": group.ID,
			})

			if err != nil {
				fmt.Println("Failed to delete group " + group.Name)
				fmt.Println(err)
				continue
			}
		} else {
			repoRemoveText := ""
			if group.AutoRemoveUsers {
				repoRemoveText = " and have been removed from repo " + group.GitHubRepo
			}

			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      group.Belongs,
				Notification: "Group " + group.Name + " has been triggered auto remove on " + fmt.Sprintf("%d", len(users)) + " users" + repoRemoveText + "!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Group Auto Remove",
				UserGroup:    group.ID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "warning",
			})

			_, err = database.MongoDB.Collection("userGroup").UpdateOne(context.Background(), bson.M{
				"_id": group.ID,
			}, bson.M{
				"$set": bson.M{
					"notified": true,
				},
			})
		}
	}
}

/*
*
idId:
* 0 = userGroup
* 1 = gitHubUser
* 2 = token
*/
func sendErrorNotification(title string, notification string, belongs primitive.ObjectID, idId int, id primitive.ObjectID) {
	userGroup := primitive.NilObjectID
	gitHubUser := primitive.NilObjectID
	token := primitive.NilObjectID

	switch idId {
	case 0:
		userGroup = id
		break
	case 1:
		gitHubUser = id
		break
	case 2:
		token = id
	}

	database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
		ID:           primitive.NewObjectID(),
		Belongs:      belongs,
		Notification: notification,
		DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
		Title:        title,
		UserGroup:    userGroup,
		GitHubUser:   gitHubUser,
		Token:        token,
		Style:        "danger",
	})
}
