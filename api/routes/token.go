package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/apierr"
	"github.com/shellhub-io/shellhub/api/token"
)

const (
	GetTokenURL         = "/token/:namespace"
	CreateTokenURL      = "/token/:namespace"
	DeleteTokenURL      = "/token/:namespace"
	ChangePermissionURL = "/token/:namespace"
)

func GetToken(c apicontext.Context) error {
	token, err := token.NewService(c.Store()).GetToken(c.Ctx(), c.Param("namespace"))
	if err != nil {
		return apierr.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, token)
}

func CreateToken(c apicontext.Context) error {
	token, err := token.NewService(c.Store()).CreateToken(c.Ctx(), c.Param("namespace"))
	if err != nil {
		return apierr.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, token)
}

func DeleteToken(c apicontext.Context) error {
	if err := token.NewService(c.Store()).DeleteToken(c.Ctx(), c.Param("namespace")); err != nil {
		return apierr.HandleError(c, err)
	}

	return nil
}

func ChangePermission(c apicontext.Context) error {
	if err := token.NewService(c.Store()).ChangePermission(c.Ctx(), c.Param("namespace")); err != nil {
		return apierr.HandleError(c, err)
	}

	return nil
}
