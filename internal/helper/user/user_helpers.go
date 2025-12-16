package helper

import (
	"errors"

	"github.com/parlorhub/api-core/internal/database/sqlc"
)

func ParseUserRole(stringRole string, def sqlc.UserRole) sqlc.UserRole {
	switch stringRole {
	case "admin":
		return sqlc.UserRoleAdmin
	case "owner":
		return sqlc.UserRoleOwner
	case "employee":
		return sqlc.UserRoleEmployee
	case "customer":
		return sqlc.UserRoleCustomer
	default:
		return def
	}
}

func userRoleToInteger(userRole sqlc.UserRole) (int, error) {
	switch userRole {
	case sqlc.UserRoleCustomer:
		return 0, nil
	case sqlc.UserRoleEmployee:
		return 1, nil
	case sqlc.UserRoleOwner:
		return 2, nil
	case sqlc.UserRoleAdmin:
		return 3, nil
	default:
		return 0, errors.New("role not found")
	}
}

func CompareUserRolePrivileges(userRole1 sqlc.UserRole, userRole2 sqlc.UserRole) (bool, error) {
	role1integer, err := userRoleToInteger(userRole1)
	if err != nil {
		return false, err
	}

	role2integer, err := userRoleToInteger(userRole2)
	if err != nil {
		return false, err
	}

	return role1integer >= role2integer, nil
}

// IsOnboardingCompleted returns true if onboarding_completed_at is set OR user is at/past step 4 (last step)
const LastOnboardingStep = 4

func IsOnboardingCompleted(user sqlc.User) bool {
	if user.OnboardingCompletedAt.Valid {
		return true
	}
	if user.CurrentOnboardingStepID.Valid && user.CurrentOnboardingStepID.Int32 >= LastOnboardingStep {
		return true
	}
	return false
}
