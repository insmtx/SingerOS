package service

import (
	"context"
	"errors"

	"github.com/insmtx/SingerOS/backend/internal/api/auth"
)

func getOrgIDFromContext(ctx context.Context) (uint, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.OrgID == 0 {
		return 0, errors.New("user not authenticated or org not set")
	}
	return caller.OrgID, nil
}

func verifyOrgPermission(daOrgID, orgID uint) error {
	if daOrgID != orgID {
		return errors.New("permission denied")
	}
	return nil
}
