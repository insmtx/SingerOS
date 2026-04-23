package githubprovider

import (
	"context"
)

type clientResolver interface {
	Resolve(ctx context.Context, factory *ClientFactory, req *ResolveClientRequest) (*ResolvedClient, bool, error)
}

type installationClientResolver struct{}

func (r *installationClientResolver) Resolve(ctx context.Context, factory *ClientFactory, req *ResolveClientRequest) (*ResolvedClient, bool, error) {
	installationID := selectorExternalRef(req.Selector, "github.installation_id")
	if explicitProfileID(req) != "" || installationID == "" {
		return nil, false, nil
	}
	if !isInstallationConfigured(factory.cfg) {
		return nil, false, nil
	}

	resolved, err := factory.resolveInstallationClient(ctx, installationID)
	if err != nil {
		return nil, true, err
	}
	return resolved, true, nil
}

type oauthClientResolver struct{}

func (r *oauthClientResolver) Resolve(ctx context.Context, factory *ClientFactory, req *ResolveClientRequest) (*ResolvedClient, bool, error) {
	if factory.authService == nil {
		return nil, true, errAuthServiceRequired
	}

	resolved, err := factory.authService.ResolveAuthorization(ctx, factory.buildResolveAuthorizationRequest(req))
	if err != nil {
		return nil, true, err
	}
	if resolved.Credential.AccessToken == "" {
		return nil, true, errEmptyAccessToken
	}

	return &ResolvedClient{
		Client:     factory.newGitHubClient(resolved.Credential.AccessToken),
		Account:    resolved.Account,
		Credential: resolved.Credential,
		ResolvedBy: resolved.ResolvedBy,
	}, true, nil
}

func defaultResolvers() []clientResolver {
	return []clientResolver{
		&installationClientResolver{},
		&oauthClientResolver{},
	}
}
