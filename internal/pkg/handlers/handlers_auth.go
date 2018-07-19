package handlers

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
	"github.com/traversals/blaster/internal/pkg/config"
	"github.com/traversals/blaster/internal/pkg/utils"
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
		"chat:write:user",
	}
)

// AuthInitiate handles /auth/initiate.
func AuthInitiate(c *gin.Context) {
	redirectURI, err := authorizeURI(utils.RelativeURI(c, "/auth/complete"))
	if err != nil {
		log.Fatal(err)
	}
	c.Redirect(http.StatusSeeOther, redirectURI)
}

// AuthComplete handles /auth/complete.
func AuthComplete(c *gin.Context) {
	code := c.Query("code")

	if code != "" {
		response, err := slack.GetOAuthResponse(slackClientID, slackClientSecret, code, utils.RelativeURI(c, "/auth/complete"), false)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		setAuthorizedToken(c, response.AccessToken)
	}

	c.Redirect(http.StatusSeeOther, utils.RelativeURI(c, "/"))
}

// AuthLogout handles /auth/logout.
func AuthLogout(c *gin.Context) {
	setAuthorizedToken(c, "")

	c.Redirect(http.StatusSeeOther, utils.RelativeURI(c, "/"))
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
	c.SetCookie(config.CookiePrefix+"slacktoken", token, 86400, "", "", !config.IsDebugging, true)
}

func authorizedToken(c *gin.Context) string {
	token, _ := c.Cookie(config.CookiePrefix + "slacktoken")
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
