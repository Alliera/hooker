package git

type Body struct {
	Ref        string     `json:"ref"`
	BaseRef    *string    `json:"base_ref"`
	After      string     `json:"after"`
	HeadCommit HeadCommit `json:"head_commit"`
	Repository Repository `json:"repository"`
	Pusher     User       `json:"pusher"`
}

type Repository struct {
	Name string `json:"name"`
}

type HeadCommit struct {
	Id        string `json:"id"`
	Author    User   `json:"author"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
