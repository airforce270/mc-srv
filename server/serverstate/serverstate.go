// Package serverstate contains an enum for the current server state.
package serverstate

type State uint8

const (
	PreHandshake State = iota
	ClientRequestingStatus
	ClientRequestingLogin
	EncryptionRequested
	LoginSucceededPendingConfirmation
	LoginSucceeded
	LoginCompletePendingAcknowledgement
	LoginComplete
	ConfigurationCompletePendingAcknowledgement
	ConfigurationComplete
)
