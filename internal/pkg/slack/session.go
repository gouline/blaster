package slack

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gouline/blaster/internal/pkg/scache"
	"github.com/slack-go/slack"
)

var (
	scopes = []string{
		"team:read",
		"users:read",
		"usergroups:read",
		"im:write",
		"chat:write:bot",
		"chat:write:user",
	}

	destinationCache = scache.New(5*time.Minute, 10*time.Minute)
)

type Session interface {
	Marshal() string
	Unmarshal(data string)
	TeamName() string
	IsAuthenticated() bool
	Reset()
	Authenticate(clientID, clientSecret, redirectURI string, query url.Values) (bool, error)
	AuthorizeURL(clientID, redirectURI string) (string, error)
	GetDestinations() ([]*Destination, error)
	PostMessage(user, message string, asUser bool) error
}

type ClientSession struct {
	Token string `json:"token"`
	Team  string `json:"team"`
}

func NewSession() Session {
	return &ClientSession{}
}

func (s *ClientSession) Marshal() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (s *ClientSession) Unmarshal(data string) {
	json.Unmarshal([]byte(data), s)
}

func (s *ClientSession) TeamName() string {
	return s.Team
}

// IsAuthenticated returns true if sessions has a token.
func (s *ClientSession) IsAuthenticated() bool {
	return s.Token != ""
}

func (s *ClientSession) Reset() {
	s.Token = ""
	s.Team = ""
}

// client creates a new [slack.Client] from token.
func (s *ClientSession) client() *slack.Client {
	return slack.New(s.Token)
}

// tokenHash returns token hashed with SHA-1.
func (s *ClientSession) tokenHash() string {
	h := sha1.New()
	h.Write([]byte(s.Token))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Authenticate establishes a new session against Slack API.
func (s *ClientSession) Authenticate(clientID, clientSecret, redirectURI string, query url.Values) (bool, error) {
	codes, ok := query["code"]
	if !ok && len(codes) != 1 {
		return false, nil
	}

	response, err := slack.GetOAuthResponse(http.DefaultClient, clientID, clientSecret, codes[0], redirectURI)
	if err != nil {
		return true, err
	}
	s.Token = response.AccessToken

	teamInfo, err := s.client().GetTeamInfo()
	if err != nil {
		s.Token = ""
		return true, err
	}
	s.Team = teamInfo.Name

	return true, nil
}

func (s *ClientSession) AuthorizeURL(clientID, redirectURI string) (string, error) {
	authorizeURL, err := url.Parse("https://slack.com/oauth/authorize")
	if err != nil {
		return "", fmt.Errorf("failed to parse authorize URL: %w", err)
	}

	q := authorizeURL.Query()
	q.Set("client_id", clientID)
	q.Set("scope", strings.Join(scopes, ","))
	q.Set("redirect_uri", redirectURI)
	authorizeURL.RawQuery = q.Encode()

	return authorizeURL.String(), nil
}

// Destionation represents user or user group.
type Destination struct {
	Type        string
	Name        string
	DisplayName string
	ID          string
	Children    []*Destination
}

// GetDestinations retrieves a list of users and user groups that you can send messages to.
func (s *ClientSession) GetDestinations() ([]*Destination, error) {
	cacheResponse := <-destinationCache.ResponseChan(s.tokenHash(), func(key string) (interface{}, error) {
		userLookup := map[string]*Destination{}
		destinations := []*Destination{}

		users, err := s.client().GetUsers()
		if err != nil {
			return destinations, fmt.Errorf("failed to get users: %w", err)
		}
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

		usergroups, err := s.client().GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(true))
		if err != nil {
			return nil, fmt.Errorf("failed to get user groups: %w", err)
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
	if cacheResponse.Error != nil {
		return []*Destination{}, cacheResponse.Error
	}
	return cacheResponse.Value.([]*Destination), nil
}

// SendMessage sends text message to a user by ID.
// Depending on asUser, message will be sent as your authenticated user or as the app's bot.
func (s *ClientSession) PostMessage(user, message string, asUser bool) error {
	client := s.client()

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
