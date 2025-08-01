package authconstants

// Context keys for storing auth-related data
const (
	// ContextKeyUserID stores the authenticated user's ID in context
	ContextKeyUserID = "user_id"
	
	// ContextKeyAccount stores the authenticated user's account in context
	ContextKeyAccount = "account"
	
	// ContextKeySession stores the current session in context
	ContextKeySession = "session"
	
	// ContextKeyIsAuthenticated indicates if the user is authenticated
	ContextKeyIsAuthenticated = "is_authenticated"
)