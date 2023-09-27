package usersession

import (
	"context"

	"github.com/nedpals/supabase-go"

	"github.com/Sinbad-HQ/kyc/core/components/usersession/models"
)

const AuthCtxKey = "authCtx"

type Component interface {
	GetAuthContextByAccessToken(authToken string) (models.AuthContext, error)
	GetAuthContextFromCtx(ctx context.Context) *models.AuthContext
}

type component struct {
	sbClient *supabase.Client
}

func NewComponent(sbClient *supabase.Client) Component {
	return &component{
		sbClient: sbClient,
	}
}

func (c *component) GetAuthContextByAccessToken(authToken string) (models.AuthContext, error) {
	// TODO: verify access token expiration + validity
	user, err := c.sbClient.Auth.User(context.Background(), authToken)
	if err != nil {
		return models.AuthContext{}, err
	}

	return models.AuthContext{
		ProviderID: user.ID,
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
