package domain

import "trisse-home/internal/shared/kernel"

// Domain errors of the auth module.
var (
	ErrAccountNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "No account has been set up yet.")
	ErrAlreadySetUp       = kernel.NewPublicError(kernel.ErrConflict, "The account has already been set up.")
	ErrSetupRequired      = kernel.NewPublicError(kernel.ErrForbidden, "The account must be set up first.")
	ErrInvalidSetupToken  = kernel.NewPublicError(kernel.ErrUnauthorized, "The setup token is invalid.")
	ErrInvalidCredentials = kernel.NewPublicError(kernel.ErrUnauthorized, "The password is incorrect.")
	ErrSessionInvalid     = kernel.NewPublicError(kernel.ErrUnauthorized, "Your session has expired. Please log in again.")
	ErrSessionNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "Session not found.")
	ErrSessionNotListed   = kernel.NewPublicError(kernel.ErrNotFound, "The session does not exist (anymore).")
	ErrTooManyAttempts    = kernel.NewPublicError(kernel.ErrTooManyRequests, "Too many failed attempts. Please wait a few minutes.")
)

// Errors of guest accounts.
var (
	ErrOwnerNotAGuest  = kernel.NewPublicError(kernel.ErrInvalid, "The owner account is not a guest.")
	ErrGuestNotFound   = kernel.NewPublicError(kernel.ErrNotFound, "The guest does not exist.")
	ErrGuestNameTaken  = kernel.NewPublicError(kernel.ErrConflict, "A guest with this name already exists.")
	ErrTooManyGuests   = kernel.NewPublicError(kernel.ErrInvalid, "There are already 20 guests.")
	ErrAccountDisabled = kernel.NewPublicError(kernel.ErrForbidden, "This account is disabled.")
	ErrOwnerOnly       = kernel.NewPublicError(kernel.ErrForbidden, "Only an admin can do this.")
	ErrReadOnly        = kernel.NewPublicError(kernel.ErrForbidden, "Guests can only view this dashboard.")
)

// Errors of members and groups (#294).
var (
	ErrMemberNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The member does not exist.")
	// ErrSecondFactorRequired answers an account without a second factor while an admin policy requires one (#304).
	ErrSecondFactorRequired = kernel.NewPublicError(kernel.ErrForbidden, "Set up a second factor (an authenticator app or a passkey) to continue; your administrator requires one.")
	ErrTooManyMembers       = kernel.NewPublicError(kernel.ErrInvalid, "There are already 50 members.")
	ErrMembersUnavailable   = kernel.NewPublicError(kernel.ErrConflict, "Members are not available yet: personal content is not separated per account in this version.")
	ErrMemberNotAllowed     = kernel.NewPublicError(kernel.ErrForbidden, "Members cannot change this; ask an admin.")
	ErrGroupNotFound        = kernel.NewPublicError(kernel.ErrNotFound, "The group does not exist.")
	ErrGroupNameTaken       = kernel.NewPublicError(kernel.ErrConflict, "A group with this name already exists.")
	ErrTooManyGroups        = kernel.NewPublicError(kernel.ErrInvalid, "There are already 50 groups.")
)

// Errors of API tokens.
var (
	ErrTokenNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The token does not exist.")
	ErrTooManyTokens = kernel.NewPublicError(kernel.ErrInvalid, "There are already 50 tokens.")
	ErrTokenScope    = kernel.NewPublicError(kernel.ErrForbidden, "The token is not allowed to do this.")
)

// Errors of single sign-on.
var (
	ErrSSONotAllowed       = kernel.NewPublicError(kernel.ErrForbidden, "This account is not allowed to sign in here.")
	ErrSSOStateInvalid     = kernel.NewPublicError(kernel.ErrUnauthorized, "The sign-in took too long or was started elsewhere. Please try again.")
	ErrSSOUnavailable      = kernel.NewPublicError(kernel.ErrNotFound, "Single sign-on is not configured.")
	ErrLocalLoginDisabled  = kernel.NewPublicError(kernel.ErrForbidden, "Password login is disabled. Use single sign-on.")
	ErrSSOSecondFactorNeed = kernel.NewPublicError(kernel.ErrUnauthorized, "Enter your authentication code to finish signing in.")
)
