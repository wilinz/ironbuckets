package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/damacus/iron-buckets/internal/handlers"
	"github.com/damacus/iron-buckets/internal/middleware"
	"github.com/damacus/iron-buckets/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/minio/madmin-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGroupManagementJourney(t *testing.T) {
	// 1. Setup
	e := echo.New()
	e.Renderer = &MockRenderer{}
	authService := services.NewAuthService()
	mockFactory := new(MockMinioFactory)
	mockClient := new(MockMinioClient)

	creds := services.Credentials{Endpoint: "play.minio.io:9000", AccessKey: "admin", SecretKey: "password"}
	mockFactory.On("NewAdminClient", creds).Return(mockClient, nil)

	// Mock listing groups
	mockClient.On("ListGroups", mock.Anything).Return([]string{"developers", "admins"}, nil)

	// Mock getting group description for all groups
	mockClient.On("GetGroupDescription", mock.Anything, "developers").Return(&madmin.GroupDesc{
		Name:    "developers",
		Members: []string{"alice", "bob"},
		Policy:  "readwrite",
		Status:  "enabled",
	}, nil)
	mockClient.On("GetGroupDescription", mock.Anything, "admins").Return(&madmin.GroupDesc{
		Name:    "admins",
		Members: []string{"root"},
		Policy:  "consoleAdmin",
		Status:  "enabled",
	}, nil)

	// Mock creating a group and adding members - use mock.MatchedBy for flexible matching
	mockClient.On("UpdateGroupMembers", mock.Anything, mock.MatchedBy(func(req madmin.GroupAddRemove) bool {
		return req.Group == "newgroup" && !req.IsRemove
	})).Return(nil)

	// Mock setting group status (disable)
	mockClient.On("SetGroupStatus", mock.Anything, "newgroup", madmin.GroupDisabled).Return(nil)

	// Mock attaching policy to group
	mockClient.On("SetPolicy", mock.Anything, "readonly", "newgroup", true).Return(nil)

	encrypted, _ := authService.EncryptCredentials(creds)
	cookie := &http.Cookie{Name: "IronSeal", Value: encrypted}

	// Routes
	app := e.Group("")
	app.Use(middleware.AuthMiddleware(authService))
	groupsHandler := handlers.NewGroupsHandler(mockFactory)
	app.GET("/groups", groupsHandler.ListGroups)
	app.GET("/groups/create", groupsHandler.CreateGroupModal)
	app.POST("/groups/create", groupsHandler.CreateGroup)
	app.GET("/groups/:groupName", groupsHandler.ViewGroup)
	app.POST("/groups/:groupName/members/add", groupsHandler.AddMembers)
	app.POST("/groups/:groupName/members/remove", groupsHandler.RemoveMembers)
	app.POST("/groups/:groupName/disable", groupsHandler.DisableGroup)
	app.POST("/groups/:groupName/enable", groupsHandler.EnableGroup)
	app.POST("/groups/:groupName/policy", groupsHandler.AttachPolicy)

	// Step A: List Groups
	reqList := httptest.NewRequest(http.MethodGet, "/groups", nil)
	reqList.AddCookie(cookie)
	recList := httptest.NewRecorder()
	e.ServeHTTP(recList, reqList)
	assert.Equal(t, http.StatusOK, recList.Code)

	// Step B: Create Empty Group
	formCreate := make(url.Values)
	formCreate.Set("groupName", "newgroup")

	reqCreate := httptest.NewRequest(http.MethodPost, "/groups/create", strings.NewReader(formCreate.Encode()))
	reqCreate.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqCreate.AddCookie(cookie)
	recCreate := httptest.NewRecorder()
	e.ServeHTTP(recCreate, reqCreate)
	assert.Equal(t, http.StatusOK, recCreate.Code)
	assert.Equal(t, "/groups", recCreate.Header().Get("HX-Redirect"))

	// Step C: Add Member to Group
	formAddMember := make(url.Values)
	formAddMember.Set("members", "alice")

	reqAddMember := httptest.NewRequest(http.MethodPost, "/groups/newgroup/members/add", strings.NewReader(formAddMember.Encode()))
	reqAddMember.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqAddMember.AddCookie(cookie)
	recAddMember := httptest.NewRecorder()
	e.ServeHTTP(recAddMember, reqAddMember)
	assert.Equal(t, http.StatusOK, recAddMember.Code)

	// Step D: Attach Policy to Group
	formPolicy := make(url.Values)
	formPolicy.Set("policy", "readonly")

	reqPolicy := httptest.NewRequest(http.MethodPost, "/groups/newgroup/policy", strings.NewReader(formPolicy.Encode()))
	reqPolicy.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqPolicy.AddCookie(cookie)
	recPolicy := httptest.NewRecorder()
	e.ServeHTTP(recPolicy, reqPolicy)
	assert.Equal(t, http.StatusOK, recPolicy.Code)

	// Step E: Disable Group
	reqDisable := httptest.NewRequest(http.MethodPost, "/groups/newgroup/disable", nil)
	reqDisable.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqDisable.AddCookie(cookie)
	recDisable := httptest.NewRecorder()
	e.ServeHTTP(recDisable, reqDisable)
	assert.Equal(t, http.StatusOK, recDisable.Code)
}
