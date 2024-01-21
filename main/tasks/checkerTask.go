package tasks

import (
	"ASO/main/database"
	"ASO/main/database/models"
	"ASO/main/git"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

func initCheckerTasks() {
	_, err := Scheduler.Every(1).Hours().Do(func() {
		checkTokens()
		checkGitUsers()
		checkGroups()
		checkSoonExpireGroups()
		checkUserGitState()
	})

	if err != nil {
		fmt.Println("Failed to start checker task")
		fmt.Println(err)
	}
}

func checkUserGitState() {
	curr, err := database.MongoDB.Collection("githubUser").Find(context.Background(), bson.M{
		"userGroup": bson.M{
			"$ne": primitive.NilObjectID,
		},
		"addedToRepo": false,
	})

	if err != nil {
		fmt.Println("Failed to get users")
		fmt.Println(err)
		return
	}

	var users []models.GitHubUser
	err = curr.All(context.Background(), &users)

	if err != nil {
		fmt.Println("Failed to parse users")
		fmt.Println(err)
		return
	}

	cachedAdmins := map[string]models.User{}
	cachedGroups := map[string]models.UserGroup{}

	for _, user := range users {
		grp := cachedGroups[user.UserGroup.Hex()]
		if grp.ID == primitive.NilObjectID {
			err = database.MongoDB.Collection("userGroup").FindOne(context.Background(), bson.M{
				"_id": user.UserGroup,
			}).Decode(&grp)

			if err != nil {
				fmt.Println("Failed to get group")
				fmt.Println(err)
				continue
			}

			cachedGroups[user.UserGroup.Hex()] = grp
		}

		admin := cachedAdmins[grp.Belongs.Hex()]
		if admin.ID == primitive.NilObjectID {
			err = database.MongoDB.Collection("user").FindOne(context.Background(), bson.M{
				"_id": grp.Belongs,
			}).Decode(&admin)

			if err != nil {
				fmt.Println("Failed to get admin")
				fmt.Println(err)
				continue
			}

			cachedAdmins[grp.Belongs.Hex()] = admin
		}

		if git.CheckIfUserIsColabo(grp.GitHubOwner, user.GitHubUsername, admin.GitHubToken, grp.GitHubRepo) {
			_, err = database.MongoDB.Collection("githubUser").UpdateOne(context.Background(), bson.M{
				"_id": user.ID,
			}, bson.M{
				"$set": bson.M{
					"addedToRepo": true,
				},
			})

			if err != nil {
				fmt.Println("Failed to update user")
				fmt.Println(err)
				continue
			}

			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      user.Belongs,
				Notification: "User " + user.Username + " has been added to repo " + grp.GitHubRepo + " by external source",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "User added to repo",
				UserGroup:    grp.ID,
				GitHubUser:   user.ID,
				Token:        primitive.NilObjectID,
				Style:        "success",
			})
		} else if git.CheckIfUserIsPendingInvite(grp.GitHubOwner, user.GitHubUsername, admin.GitHubToken, grp.GitHubRepo) {
			_, err = database.MongoDB.Collection("githubUser").UpdateOne(context.Background(), bson.M{
				"_id": user.ID,
			}, bson.M{
				"$set": bson.M{
					"addedToRepo": true,
				},
			})

			if err != nil {
				fmt.Println("Failed to update user")
				fmt.Println(err)
				continue
			}

			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      user.Belongs,
				Notification: "User " + user.Username + " has been added to repo " + grp.GitHubRepo + " by external source",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "User added to repo",
				UserGroup:    grp.ID,
				GitHubUser:   user.ID,
				Token:        primitive.NilObjectID,
				Style:        "success",
			})
		}
	}
}

func checkSoonExpireGroups() {
	fmt.Println("Checking soon expire groups")
	cur, err := database.MongoDB.Collection("userGroup").Find(context.Background(), bson.M{
		"expires":         true,
		"notify":          true,
		"notifiedExpired": false,
		"notifiedDeleted": false,
		"dateExpires": bson.M{
			"$lte": time.Now().Add(time.Hour * 24 * 30 * -1),
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
		fmt.Println("Group soon expired: " + group.Name)

		//get days till expire
		toTime := group.DateExpires.Time()
		now := time.Now()
		days := int(toTime.Sub(now).Hours() / 24)
		daysText := strconv.Itoa(days)

		if days == 1 {
			daysText = "1 day"
		} else if days == 0 {
			daysText = "today"
		} else {
			daysText += " days"
		}

		database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
			ID:           primitive.NewObjectID(),
			Belongs:      group.Belongs,
			Notification: "Group " + group.Name + " will expire in " + daysText + "!",
			DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
			Title:        "Group " + group.Name + " will soon expired",
			Style:        "warning",
			UserGroup:    group.ID,
		})

		_, err = database.MongoDB.Collection("userGroup").UpdateOne(context.Background(), bson.M{
			"_id": group.ID,
		}, bson.M{
			"$set": bson.M{
				"notifiedExpired": true,
			},
		})
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
				Notification: "Token " + token.Name + " has been deleted due to expiration!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Token " + token.Name + " deleted",
				UserGroup:    primitive.NilObjectID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "info",
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

		removedRepo := false
		groupExists := false

		if err == nil {
			if user.AddedToRepo {
				if !git.RemoveUserFromRepo(creator.GitHubUsername, user.GitHubUsername, creator.GitHubToken, user.UserGroup.Hex()) {
					fmt.Println("Failed to remove user " + user.Username + " from repo " + user.UserGroup.Hex())
					sendErrorNotification("Failed to remove user from repo", "Failed to remove user "+user.Username+" from repo "+group.GitHubRepo, user.Belongs, 1, user.ID)
					continue
				}
				removedRepo = true
			}
			groupExists = true
		}

		_, err = database.MongoDB.Collection("githubUser").DeleteOne(context.Background(), bson.M{
			"_id": user.ID,
		})

		if err != nil {
			fmt.Println("Failed to delete user " + user.Username)
			fmt.Println(err)
			continue
		}

		extraText := ""

		if groupExists {
			extraText = "<br>User was removed from group " + group.Name
		}

		if removedRepo {
			extraText = "<br>User was removed from repo " + group.GitHubRepo
		}

		database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
			ID:           primitive.NewObjectID(),
			Belongs:      user.Belongs,
			Notification: "User " + user.Username + " has been deleted due to expiration!" + extraText,
			DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
			Title:        "User " + user.Username + " deleted",
			UserGroup:    group.ID,
			GitHubUser:   primitive.NilObjectID,
			Token:        primitive.NilObjectID,
			Style:        "info",
		})
	}
}

func checkGroups() {
	fmt.Println("Checking groups")
	cur, err := database.MongoDB.Collection("userGroup").Find(context.Background(), bson.M{
		"expires":         true,
		"notifiedDeleted": false,
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

		skipped := []string{}

		if group.AutoRemoveUsers || group.AutoDelete {
			for _, user := range users {
				if user.ExpiresGroup && (group.AutoDelete || group.AutoRemoveUsers) {
					if user.AddedToRepo && group.AutoRemoveUsers {
						if !git.RemoveUserFromRepo(group.GitHubOwner, user.GitHubUsername, creator.GitHubToken, group.GitHubRepo) {
							fmt.Println("Failed to remove user " + user.Username + " from repo " + group.GitHubRepo)
							sendErrorNotification("Failed to remove user from repo", "Failed to remove user "+user.Username+" from repo "+group.GitHubRepo, user.Belongs, 1, user.ID)
							continue
						}
					}

					if group.AutoDelete {
						_, err = database.MongoDB.Collection("gitUser").DeleteOne(context.Background(), bson.M{
							"_id": user.ID,
						})
					}
				} else {
					skipped = append(skipped, user.Username)

					if group.AutoDelete {
						_, err = database.MongoDB.Collection("gitUser").UpdateOne(context.Background(), bson.M{
							"_id": user.ID,
						}, bson.M{
							"$set": bson.M{
								"userGroup": primitive.NilObjectID,
							},
						})
					}

					groupId := primitive.NilObjectID

					if !group.AutoDelete {
						groupId = group.ID
					}

					database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
						ID:           primitive.NewObjectID(),
						Belongs:      group.Belongs,
						Notification: "User " + user.Username + " has <b>not</b> been removed from group " + group.Name + ", because it doesn't expire by group!",
						DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
						Title:        "User not removed from repo",
						UserGroup:    groupId,
						GitHubUser:   user.ID,
						Token:        primitive.NilObjectID,
						Style:        "warning",
					})
				}
			}
		}

		if group.AutoDelete {
			repoRemoveText := ""
			if group.AutoRemoveUsers {
				repoRemoveText = " and removed from repo " + group.GitHubRepo
			}

			if len(skipped) > 0 {
				repoRemoveText += ", skipped " + fmt.Sprintf("%d", len(skipped)) + " users because they are not expired by group"
			}

			_, err = database.MongoDB.Collection("userGroup").DeleteOne(context.Background(), bson.M{
				"_id": group.ID,
			})

			if err != nil {
				fmt.Println("Failed to delete group " + group.Name)
				fmt.Println(err)
				continue
			}

			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      group.Belongs,
				Notification: "Group " + group.Name + " has been deleted! " + fmt.Sprintf("%d", len(users)-len(skipped)) + " users have been removed from the group" + repoRemoveText + "!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Group " + group.Name + " triggered Auto Delete",
				UserGroup:    primitive.NilObjectID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "info",
			})
		} else if group.AutoRemoveUsers {
			repoRemoveText := ""

			if len(skipped) > 0 {
				repoRemoveText += ", skipped " + fmt.Sprintf("%d", len(skipped)) + " users because they are not expired by group"
			}

			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      group.Belongs,
				Notification: "Group " + group.Name + " has triggered auto remove on " + fmt.Sprintf("%d", len(users)-len(skipped)) + " users. They have been removed from repo " + group.GitHubRepo + repoRemoveText + "!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Group " + group.Name + " has triggered Auto Remove",
				UserGroup:    group.ID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "info",
			})

			_, err = database.MongoDB.Collection("userGroup").UpdateOne(context.Background(), bson.M{
				"_id": group.ID,
			}, bson.M{
				"$set": bson.M{
					"notifiedDeleted": true,
					"notifiedExpired": true,
				},
			})
		} else {
			database.MongoDB.Collection("notification").InsertOne(context.Background(), models.Notification{
				ID:           primitive.NewObjectID(),
				Belongs:      group.Belongs,
				Notification: "Group " + group.Name + " has expired! No actions have been taken because auto delete and auto remove are disabled!",
				DateCreated:  primitive.NewDateTimeFromTime(time.Now()),
				Title:        "Group " + group.Name + " has expired",
				UserGroup:    group.ID,
				GitHubUser:   primitive.NilObjectID,
				Token:        primitive.NilObjectID,
				Style:        "warning",
			})

			_, err = database.MongoDB.Collection("userGroup").UpdateOne(context.Background(), bson.M{
				"_id": group.ID,
			}, bson.M{
				"$set": bson.M{
					"notifiedDeleted": true,
					"notifiedExpired": true,
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
