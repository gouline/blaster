package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mgouline/slack"
)

const slackBaseURL = "https://slack.com"

var (
	slackClientID     = os.Getenv("SLACK_CLIENT_ID")
	slackClientSecret = os.Getenv("SLACK_CLIENT_SECRET")

	slackAPIScopes = []string{
		"team:read",
		"users:read",
		"usergroups:read",
		"im:write",
		"chat:write:bot",
	}
)

func handleAuthInitiate(c *gin.Context) {
	redirectURI, err := authorizeURI(relativeURI(c, "/auth/complete"))
	if err != nil {
		log.Fatal(err)
	}
	c.Redirect(http.StatusSeeOther, redirectURI)
}

func handleAuthComplete(c *gin.Context) {
	code := c.Query("code")

	if code != "" {
		response, err := slack.GetOAuthResponse(slackClientID, slackClientSecret, code, relativeURI(c, "/auth/complete"), false)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		setAuthorizedToken(c, response.AccessToken)
	}

	c.Redirect(http.StatusSeeOther, relativeURI(c, "/"))
}

func handleAuthLogout(c *gin.Context) {
	setAuthorizedToken(c, "")

	c.Redirect(http.StatusSeeOther, relativeURI(c, "/"))
}

func authorizeURI(redirectURI string) (string, error) {
	redirectURL, err := url.Parse(slackBaseURL + "/oauth/authorize")
	if err != nil {
		return "", err
	}
	q := redirectURL.Query()
	q.Set("client_id", slackClientID)
	q.Set("scope", strings.Join(slackAPIScopes, ","))
	q.Set("redirect_uri", redirectURI)
	redirectURL.RawQuery = q.Encode()

	return redirectURL.String(), nil
}

func setAuthorizedToken(c *gin.Context, token string) {
	c.SetCookie(cookiePrefix+"slacktoken", token, 86400, "", "", !isDebugging, true)
}

func authorizedToken(c *gin.Context) string {
	token, _ := c.Cookie(cookiePrefix + "slacktoken")
	return token
}

func isAuthorized(c *gin.Context) bool {
	return authorizedToken(c) != ""
}

func hashedToken(token string) string {
	h := sha1.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum(nil))
}
