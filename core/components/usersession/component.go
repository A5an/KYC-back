package usersession

import (
	"context"
	"errors"

	"github.com/clerkinc/clerk-sdk-go/clerk"

	"github.com/Sinbad-HQ/kyc/core/components/usersession/models"
)

const (
	AuthCtxKey = "authCtx"

	AdminRole  = "admin"
	MemberRole = "basic_member"
)

type Component interface {
	GetAuthContextByAccessToken(authToken string) (models.AuthContext, error)
	GetAuthContextFromCtx(ctx context.Context) *models.AuthContext
	GetOrgMembers(ctx context.Context) ([]models.OrgMember, error)
	AddOrgMember(ctx context.Context, email string, role string) error
	RemoveOrgMember(ctx context.Context, userID string) error
}

type component struct {
	client clerk.Client
}

func NewComponent(clerkClient clerk.Client) Component {
	return &component{
		client: clerkClient,
	}
}

func (c *component) GetAuthContextByAccessToken(authToken string) (models.AuthContext, error) {
	sessClaims, err := c.client.VerifyToken(authToken)
	if err != nil {
		return models.AuthContext{}, errors.New("unauthorized")
	}

	user, err := c.client.Users().Read(sessClaims.Claims.Subject)
	if err != nil {
		return models.AuthContext{}, err
	}

	memberships, err := c.client.Users().ListMemberships(clerk.ListMembershipsParams{UserID: user.ID})
	if err != nil {
		return models.AuthContext{}, err
	}
	if len(memberships.Data) < 0 {
		return models.AuthContext{}, errors.New("user has no organization membership")
	}

	var orgID string
	// Users are enforced to a single organization
	if org := memberships.Data[0].Organization; org != nil {
		orgID = org.ID
	} else {
		return models.AuthContext{}, errors.New("user has no organization membership")
	}

	return models.AuthContext{
		OrgID:  orgID,
		UserID: user.ID,
		Role:   memberships.Data[0].Role,
	}, nil
}

func (c *component) GetAuthContextFromCtx(ctx context.Context) *models.AuthContext {
	if ctx == nil {
		return nil
	}
	value := ctx.Value(AuthCtxKey)
	if value == nil {
		return nil
	}
	authContext, ok := value.(*models.AuthContext)
	if !ok {
		return nil
	}

	return authContext
}

func (c *component) GetOrgMembers(ctx context.Context) ([]models.OrgMember, error) {
	authCtx := c.GetAuthContextFromCtx(ctx)
	memberships, err := c.client.Organizations().ListMemberships(clerk.ListOrganizationMembershipsParams{
		OrganizationID: authCtx.OrgID,
	})
	if err != nil {
		return nil, err
	}

	var members []models.OrgMember
	for _, membership := range memberships.Data {
		if userData := membership.PublicUserData; userData != nil {
			var name, imageURL string

			if userData.FirstName != nil {
				name += *userData.FirstName
			}
			if userData.LastName != nil {
				name += " " + *userData.LastName
			}

			if userData.ImageURL != nil {
				imageURL = *userData.ImageURL
			}

			members = append(members, models.OrgMember{
				Name:         name,
				ProfileImage: imageURL,
				Email:        userData.Identifier,
				Role:         membership.Role,
				UserID:       userData.UserID,
			})
		}

	}

	return members, nil
}

func (c *component) AddOrgMember(ctx context.Context, email string, role string) error {
	authCtx := c.GetAuthContextFromCtx(ctx)
	// only admin can add the org member
	if authCtx.Role != AdminRole {
		return errors.New("only admin can invite members to organization")
	}

	_, err := c.client.Organizations().CreateInvitation(clerk.CreateOrganizationInvitationParams{
		EmailAddress:   email,
		InviterUserID:  authCtx.UserID,
		OrganizationID: authCtx.OrgID,
		Role:           role,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *component) RemoveOrgMember(ctx context.Context, userID string) error {
	authCtx := c.GetAuthContextFromCtx(ctx)

	// only admin can delete the org member
	if authCtx.Role != AdminRole {
		return errors.New("only admin can invite members to organization")
	}

	// Prevent admin from removing themselves from membership
	if authCtx.UserID == userID {
		return errors.New("admin self delete is not allowed")
	}

	if _, err := c.client.Organizations().DeleteMembership(authCtx.OrgID, userID); err != nil {
		return err
	}

	if _, err := c.client.Users().Delete(userID); err != nil {
		return err
	}

	return nil
}
