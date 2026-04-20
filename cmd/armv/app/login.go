package app

import (
	"context"
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
	"github.com/logrusorgru/aurora"
)

// checkLogin verifies the caller has access to the source Azure subscription.
func checkLogin(ctx context.Context, azureResourceMoveInfo *validation.AzureResourceMoveInfo) error {
	login, err := auth.CheckLogin(ctx, azureResourceMoveInfo.Credentials, azureResourceMoveInfo.SourceSubscriptionId)
	if err != nil {
		return fmt.Errorf("login error: %w", err)
	}
	if !login {
		return fmt.Errorf("not logged into Azure subscription %q: please run `az login` and retry", azureResourceMoveInfo.SourceSubscriptionId)
	}
	fmt.Println(aurora.Yellow(fmt.Sprintf("Logged into Subscription Id: %s", azureResourceMoveInfo.SourceSubscriptionId)))

	return nil
}
