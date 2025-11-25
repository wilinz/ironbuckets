package handlers

import (
	"net/http"
	"strings"

	"github.com/damacus/iron-buckets/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/minio/madmin-go/v3"
)

type GroupsHandler struct {
	minioFactory services.MinioClientFactory
}

func NewGroupsHandler(minioFactory services.MinioClientFactory) *GroupsHandler {
	return &GroupsHandler{minioFactory: minioFactory}
}

// GroupInfo holds group data for templates
type GroupInfo struct {
	Name        string
	MemberCount int
	Members     []string
	Policy      string
	Status      string
}

// ListGroups renders the groups management page
func (h *GroupsHandler) ListGroups(c echo.Context) error {
	creds, err := GetCredentialsOrRedirect(c)
	if err != nil {
		return err
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	// Fetch group names
	groupNames, err := mdm.ListGroups(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list groups")
	}

	// Fetch details for each group
	groups := make([]GroupInfo, 0, len(groupNames))
	for _, name := range groupNames {
		desc, err := mdm.GetGroupDescription(c.Request().Context(), name)
		if err != nil {
			// Skip groups we can't get info for
			continue
		}
		groups = append(groups, GroupInfo{
			Name:        desc.Name,
			MemberCount: len(desc.Members),
			Members:     desc.Members,
			Policy:      desc.Policy,
			Status:      desc.Status,
		})
	}

	return c.Render(http.StatusOK, "groups", map[string]interface{}{
		"ActiveNav": "groups",
		"Groups":    groups,
	})
}

// CreateGroupModal renders the modal form for creating a group
func (h *GroupsHandler) CreateGroupModal(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	// Fetch users for the member selection
	users, err := mdm.ListUsers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users")
	}

	// Fetch policies for the policy selection
	policies, err := mdm.ListCannedPolicies(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list policies")
	}

	policyNames := make([]string, 0, len(policies))
	for name := range policies {
		policyNames = append(policyNames, name)
	}

	userNames := make([]string, 0, len(users))
	for name := range users {
		userNames = append(userNames, name)
	}

	return c.Render(http.StatusOK, "group_create_modal", map[string]interface{}{
		"Users":    userNames,
		"Policies": policyNames,
	})
}

// CreateGroup handles the creation of a new group
func (h *GroupsHandler) CreateGroup(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	groupName := c.FormValue("groupName")
	membersStr := c.FormValue("members") // comma-separated or empty
	policy := c.FormValue("policy")

	if groupName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Group name is required")
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	// Parse members (can be empty for empty group)
	var members []string
	if membersStr != "" {
		members = strings.Split(membersStr, ",")
		for i := range members {
			members[i] = strings.TrimSpace(members[i])
		}
	}

	// Create the group by adding members (empty members creates empty group)
	err = mdm.UpdateGroupMembers(c.Request().Context(), madmin.GroupAddRemove{
		Group:    groupName,
		Members:  members,
		Status:   madmin.GroupEnabled,
		IsRemove: false,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create group: "+err.Error())
	}

	// Attach policy if provided
	if policy != "" {
		if err := mdm.SetPolicy(c.Request().Context(), policy, groupName, true); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Group created but failed to assign policy: "+err.Error())
		}
	}

	return HTMXRedirect(c, "/groups")
}

// ViewGroup renders the group detail page
func (h *GroupsHandler) ViewGroup(c echo.Context) error {
	creds, err := GetCredentialsOrRedirect(c)
	if err != nil {
		return err
	}

	groupName := c.Param("groupName")

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	desc, err := mdm.GetGroupDescription(c.Request().Context(), groupName)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Group not found")
	}

	// Get all users for adding members
	users, err := mdm.ListUsers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users")
	}

	// Filter out users already in the group
	memberSet := make(map[string]bool)
	for _, m := range desc.Members {
		memberSet[m] = true
	}

	availableUsers := make([]string, 0)
	for name := range users {
		if !memberSet[name] {
			availableUsers = append(availableUsers, name)
		}
	}

	// Get policies for policy dropdown
	policies, err := mdm.ListCannedPolicies(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list policies")
	}

	policyNames := make([]string, 0, len(policies))
	for name := range policies {
		policyNames = append(policyNames, name)
	}

	return c.Render(http.StatusOK, "group_detail", map[string]interface{}{
		"ActiveNav":      "groups",
		"Group":          desc,
		"AvailableUsers": availableUsers,
		"Policies":       policyNames,
	})
}

// AddMembers adds members to a group
func (h *GroupsHandler) AddMembers(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	groupName := c.Param("groupName")
	membersStr := c.FormValue("members")

	if membersStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one member is required")
	}

	members := strings.Split(membersStr, ",")
	for i := range members {
		members[i] = strings.TrimSpace(members[i])
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	err = mdm.UpdateGroupMembers(c.Request().Context(), madmin.GroupAddRemove{
		Group:    groupName,
		Members:  members,
		Status:   madmin.GroupEnabled,
		IsRemove: false,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add members: "+err.Error())
	}

	return HTMXRedirect(c, "/groups/"+groupName)
}

// RemoveMembers removes members from a group
func (h *GroupsHandler) RemoveMembers(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	groupName := c.Param("groupName")
	membersStr := c.FormValue("members")

	if membersStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one member is required")
	}

	members := strings.Split(membersStr, ",")
	for i := range members {
		members[i] = strings.TrimSpace(members[i])
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	err = mdm.UpdateGroupMembers(c.Request().Context(), madmin.GroupAddRemove{
		Group:    groupName,
		Members:  members,
		IsRemove: true,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to remove members: "+err.Error())
	}

	return HTMXRedirect(c, "/groups/"+groupName)
}

// DisableGroup disables a group
func (h *GroupsHandler) DisableGroup(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	groupName := c.Param("groupName")

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	if err := mdm.SetGroupStatus(c.Request().Context(), groupName, madmin.GroupDisabled); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to disable group")
	}

	return HTMXRedirect(c, "/groups")
}

// EnableGroup enables a group
func (h *GroupsHandler) EnableGroup(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	groupName := c.Param("groupName")

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	if err := mdm.SetGroupStatus(c.Request().Context(), groupName, madmin.GroupEnabled); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to enable group")
	}

	return HTMXRedirect(c, "/groups")
}

// AttachPolicy attaches a policy to a group
func (h *GroupsHandler) AttachPolicy(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	groupName := c.Param("groupName")
	policy := c.FormValue("policy")

	if policy == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Policy is required")
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	// isGroup = true for group policy attachment
	if err := mdm.SetPolicy(c.Request().Context(), policy, groupName, true); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to attach policy: "+err.Error())
	}

	return HTMXRedirect(c, "/groups/"+groupName)
}
