package router

import "github.com/softbrewery/gojoi/pkg/joi"

var UsernameSchema = joi.String().Min(3).Max(30).Regex("^[a-zA-Z][a-zA-Z._\\s]+[a-zA-Z]$")
var GitHubUsername = joi.String().Min(3).Max(30).Regex("[a-zA-Z0-9-]+")
var PasswordSchema = joi.String().Min(8).Max(50).Regex("[a-zA-Z0-9._!?#\\-\\s\\*]+")
var EmailSchema = joi.String().Min(5).Max(50).Email(&joi.EmailOptions{
	SMTPLookup: false,
})
var GitTokenSchema = joi.String().Min(10).Max(150).Regex("[a-zA-Z0-9_]+")
