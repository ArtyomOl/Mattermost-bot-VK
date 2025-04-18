package storage

type Storage interface {
	CreatePoll(poll Poll) (string, error)
	GetPoll(id string) (*Poll, error)
	Vote(pollID, userID, option string) error
	GetResults(pollID string) (map[string]int, error)
	EndPoll(pollID, userID string) error
	DeletePoll(pollID, userID string) error
	GetUserPolls(userID string) ([]Poll, error)
	Close() error
}

type Poll struct {
	ID        string            `json:"id"`
	CreatorID string            `json:"creator_id"`
	Question  string            `json:"question"`
	Options   map[string]string `json:"options"`
	CreatedAt int64             `json:"created_at"`
	IsActive  bool              `json:"is_active"`
}

type Vote struct {
	PollID  string `json:"poll_id"`
	UserID  string `json:"user_id"`
	Option  string `json:"option"`
	VotedAt int64  `json:"voted_at"`
}

type PollResult struct {
	PollID     string            `json:"poll_id"`
	Question   string            `json:"question"`
	Options    map[string]string `json:"options"`
	Votes      map[string]int    `json:"votes"`
	TotalVotes int               `json:"total_votes"`
	IsActive   bool              `json:"is_active"`
}
