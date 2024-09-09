package slack

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/gouline/blaster/internal/pkg/scache"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

var (
	teamCache        = scache.New(12*time.Hour, 12*time.Hour)
	destinationCache = scache.New(5*time.Minute, 10*time.Minute)
)

// Context contains authenticated session information.
type Context struct {
	Authorized bool
	TeamName   string
	token      string
}

// Context retrieves current authentication context.
func (s *Slack) Context(c echo.Context) *Context {
	ctx := &Context{}

	if token := s.token(c); token != "" {
		ctx.Authorized = true

		cacheResponse := <-teamCache.ResponseChan(hashToken(token), func(key string) (interface{}, error) {
			client := slack.New(token)

			teamInfo, err := client.GetTeamInfo()
			if err != nil {
				return nil, err
			}

			return teamInfo.Name, err
		})
		if cacheResponse.Error == nil {
			ctx.TeamName = cacheResponse.Value.(string)
		}

		// Build other caches
		go func() {
			<-buildDestinationCache(token)
		}()

		ctx.token = token
	}

	return ctx
}

// SendMessage sends text message to a user by ID.
// Depending on asUser, message will be sent as your authenticated user or as the app's bot.
func (ctx *Context) SendMessage(user, message string, asUser bool) error {
	client := slack.New(ctx.token)

	// Open/get channel by user ID
	channel, _, _, err := client.OpenConversation(&slack.OpenConversationParameters{
		Users: []string{user},
	})
	if err != nil {
		return err
	}

	// Post message to opened channel
	_, _, err = client.PostMessage(
		channel.ID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionAsUser(asUser),
	)
	return err
}

// Destinations retrieves a list of users and user groups that you can send messages to.
func (ctx *Context) Destinations() ([]*Destination, error) {
	cacheResponse := <-buildDestinationCache(ctx.token)
	if cacheResponse.Error != nil {
		return []*Destination{}, cacheResponse.Error
	}
	return cacheResponse.Value.([]*Destination), nil
}

func buildDestinationCache(token string) <-chan scache.Response {
	return destinationCache.ResponseChan(hashToken(token), func(key string) (interface{}, error) {
		client := slack.New(token)

		var destinations []*Destination

		userLookup := map[string]*Destination{}

		// Get all users
		users, err := client.GetUsers()
		if err != nil {
			return nil, err
		}

		destinations = []*Destination{}

		for _, user := range users {
			if user.Deleted || user.IsBot {
				continue
			}

			d := &Destination{
				Type:        "user",
				Name:        user.Profile.RealName,
				DisplayName: user.Profile.DisplayName,
				ID:          user.ID,
			}

			destinations = append(destinations, d)
			userLookup[user.ID] = d
		}

		usergroups, err := client.GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(true))
		if err != nil {
			return nil, err
		}

		for _, usergroup := range usergroups {
			if !usergroup.IsUserGroup {
				continue
			}

			children := []*Destination{}

			for _, userID := range usergroup.Users {
				user, found := userLookup[userID]
				if !found {
					continue
				}

				children = append(children, user)
			}

			destinations = append(destinations, &Destination{
				Type:        "usergroup",
				Name:        usergroup.Name,
				DisplayName: usergroup.Handle,
				Children:    children,
			})
		}

		return destinations, nil
	})
}

// hashToken hashes raw auth token with SHA-1.
func hashToken(token string) string {
	h := sha1.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Destionation represents user or user group.
type Destination struct {
	Type        string
	Name        string
	DisplayName string
	ID          string
	Children    []*Destination
}
