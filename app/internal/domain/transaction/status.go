package transaction

type Status string

const (
	StatusSubmitted Status = "submitted"
	StatusMined     Status = "mined"
	StatusConfirmed Status = "confirmed"
	StatusFailed    Status = "failed"
)

func IsTerminalStatus(s Status) bool {
	return s == StatusConfirmed || s == StatusFailed
}
