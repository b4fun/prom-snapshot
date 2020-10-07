package azblob

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// CreateServincePrincipalTokenFromEnvironment creates servince principal token from automic settings.
func CreateServincePrincipalTokenFromEnvironment() (*adal.ServicePrincipalToken, error) {
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("get auth settings: %w", err)
	}

	settings.Values[auth.Resource] = settings.Environment.ResourceIdentifiers.Storage

	// 1. Client Credentials
	if c, e := settings.GetClientCredentials(); e == nil {
		return c.ServicePrincipalToken()
	}

	// 2. Client Certificate
	if c, e := settings.GetClientCertificate(); e == nil {
		return c.ServicePrincipalToken()
	}

	// 3. Username Password
	if c, e := settings.GetUsernamePassword(); e == nil {
		return c.ServicePrincipalToken()
	}

	// 4. MSI
	return settings.GetMSI().ServicePrincipalToken()
}
