package auth

// Identity represents a Kratos identity populated from JWT claims or the Kratos API.
// It is the central auth object threaded through middleware and handlers.
type Identity struct {
	// Identity is the Kratos identity UUID.
	ID string `json:"id"`
	// Traits contains the user's profile data stored in Kratos.
	Traits Traits `json:"traits"`
	// UserID is the local database user ID resolved after login.
	// It may be embedded in tokens via the "user_id" extension claim.
	UserID int `json:"-"`
}

// Traits mirrors the Kratos identity schema stored under "traits".
type Traits struct {
	Email   string  `json:"email"`
	Name    Name    `json:"name"`
	Role    string  `json:"role"`
	College College `json:"college"`
	RollNo  string  `json:"roll_no,omitempty"`
}

// Name holds the user's given and family name.
type Name struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

// College holds the college the user belongs to.
// ID is stored as a string so it can carry either a numeric ID or an
// Kratos-style external identifier such as a UUID.
type College struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// RegistrationRequest is the payload sent to Kratos to complete a
// password-based registration flow.
type RegistrationRequest struct {
	// Method must be "password" for the Kratos password strategy.
	Method   string `json:"method"`
	Password string `json:"password"`
	Traits   Traits `json:"traits"`
}
