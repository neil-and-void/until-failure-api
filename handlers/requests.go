package handlers

type (
	Verification struct {
		Attempts int    `json:"attempts"`
		ExpireAt int64  `json:"expire_at"`
		Status   string `json:"status"`
		Strategy string `json:"strategy"`
	}

	EmailAddress struct {
		EmailAddress string       `json:"email_address"`
		ID           string       `json:"id"`
		LinkedTo     []string     `json:"linked_to"`
		Object       string       `json:"object"`
		Reserved     bool         `json:"reserved"`
		Verification Verification `json:"verification"`
	}

	Data struct {
		BackupCodeEnabled             bool                   `json:"backup_code_enabled"`
		Banned                        bool                   `json:"banned"`
		CreateOrganizationEnabled     bool                   `json:"create_organization_enabled"`
		CreatedAt                     int64                  `json:"created_at"`
		DeleteSelfEnabled             bool                   `json:"delete_self_enabled"`
		EmailAddresses                []EmailAddress         `json:"email_addresses"`
		ExternalAccounts              []interface{}          `json:"external_accounts"`
		ExternalID                    interface{}            `json:"external_id"`
		FirstName                     string                 `json:"first_name"`
		HasImage                      bool                   `json:"has_image"`
		ID                            string                 `json:"id"`
		ImageURL                      string                 `json:"image_url"`
		LastActiveAt                  int64                  `json:"last_active_at"`
		LastName                      string                 `json:"last_name"`
		LastSignInAt                  interface{}            `json:"last_sign_in_at"`
		Locked                        bool                   `json:"locked"`
		LockoutExpiresInSeconds       interface{}            `json:"lockout_expires_in_seconds"`
		Object                        string                 `json:"object"`
		PasswordEnabled               bool                   `json:"password_enabled"`
		PhoneNumbers                  []interface{}          `json:"phone_numbers"`
		PrimaryEmailAddressID         string                 `json:"primary_email_address_id"`
		PrimaryPhoneNumberID          interface{}            `json:"primary_phone_number_id"`
		PrimaryWeb3WalletID           interface{}            `json:"primary_web3_wallet_id"`
		PrivateMetadata               map[string]interface{} `json:"private_metadata"`
		ProfileImageURL               string                 `json:"profile_image_url"`
		PublicMetadata                map[string]interface{} `json:"public_metadata"`
		SAMLAccounts                  []interface{}          `json:"saml_accounts"`
		TOTPEnabled                   bool                   `json:"totp_enabled"`
		TwoFactorEnabled              bool                   `json:"two_factor_enabled"`
		UnsafeMetadata                map[string]interface{} `json:"unsafe_metadata"`
		UpdatedAt                     int64                  `json:"updated_at"`
		Username                      interface{}            `json:"username"`
		VerificationAttemptsRemaining int                    `json:"verification_attempts_remaining"`
		Web3Wallets                   []interface{}          `json:"web3_wallets"`
	}

	UserCreatedEvent struct {
		Data   Data   `json:"data"`
		Object string `json:"object"`
		Type   string `json:"type"`
	}
)

// *** Requests ***
type (
	NewRoutine struct {
		Name   string `json:"name" validate:"required,min=1,max=32"`
		UserID string `json:"userId" validate:"required"`
	}

	UpdateRoutine struct {
		Name   *string `json:"name" validate:"min=1,max=32"`
		Active *bool   `json:"active"`
	}

	NewExerciseRoutine struct {
		Name      string `json:"name" validate:"required,min=1,max=99"`
		RoutineID string `json:"routineId" validate:"required"`
	}

	NewSetScheme struct {
		TargetReps        uint            `json:"targetReps" validate:"required,min=1,max=99"`
		SetType           SetType         `json:"setType" validate:"required,oneof=WARMUP WORKING DROP SUPER,required"`
		Measurement       MeasurementType `json:"measurement" validate:"required,oneof=WEIGHT DURATION BODYWEIGHT WEIGHTED_DURATION"`
		ExerciseRoutineID string          `json:"exerciseRoutineId" validate:"required"`
	}

	SetType         string
	MeasurementType string
)
