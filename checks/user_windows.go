package checks

import (
	"context"

	wapi "github.com/iamacarpet/go-win64api"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckUser) Run(ctx context.Context) (interface{}, error) {
	// This check runs best as NT AUTHORITY\SYSTEM
	//
	// Running as a normal or even elevated user,
	// we can't properly detect who is an admin or not.
	//
	// This is because we require TOKEN_DUPLICATE permission,
	// which we don't seem to have otherwise (Win10).
	users, err := wapi.ListLoggedInUsers()
	if err != nil {
		return nil, err
	}

	// Users currently logged in (Admin check doesn't work for AD Accounts)
	userResults := make([]*resultUser, 0)
	for _, u := range users {
		result := &resultUser{
			Name:     u.Username,
			Terminal: "",
			Host:     u.Domain,
			Started:  int(u.LogonTime.Unix()),
		}
		userResults = append(userResults, result)
	}

	return userResults, nil
}
