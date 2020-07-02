package api

import "time"

type paginationResponse struct {
	Page       int64   `json:"page"`
	PageLength int64   `json:"pagelen"`
	Size       int64   `json:"size"`
	Next       string  `json:"next"`
	Values     objects `json:"values"`
}

type linkResponse struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type workSpacesResponse struct {
	CreatedOn time.Time `json:"created_on"`
	IsPrivate bool      `json:"is_private"`
	Links     struct {
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Members struct {
			Href string `json:"href"`
		} `json:"members"`
		Owners struct {
			Href string `json:"href"`
		} `json:"owners"`
		Projects struct {
			Href string `json:"href"`
		} `json:"projects"`
		Repositories struct {
			Href string `json:"href"`
		} `json:"repositories"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Snippets struct {
			Href string `json:"href"`
		} `json:"snippets"`
	} `json:"links"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
	UUID string `json:"uuid"`
}

type repoResonse struct {
	CreatedOn   time.Time `json:"created_on"`
	Description string    `json:"description"`
	ForkPolicy  string    `json:"fork_policy"`
	FullName    string    `json:"full_name"`
	HasIssues   bool      `json:"has_issues"`
	HasWiki     bool      `json:"has_wiki"`
	IsPrivate   bool      `json:"is_private"`
	Language    string    `json:"language"`
	Links       struct {
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
		Branches struct {
			Href string `json:"href"`
		} `json:"branches"`
		Clone []struct {
			Href string `json:"href"`
			Name string `json:"name"`
		} `json:"clone"`
		Commits struct {
			Href string `json:"href"`
		} `json:"commits"`
		Downloads struct {
			Href string `json:"href"`
		} `json:"downloads"`
		Forks struct {
			Href string `json:"href"`
		} `json:"forks"`
		Hooks struct {
			Href string `json:"href"`
		} `json:"hooks"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Pullrequests struct {
			Href string `json:"href"`
		} `json:"pullrequests"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Source struct {
			Href string `json:"href"`
		} `json:"source"`
		Tags struct {
			Href string `json:"href"`
		} `json:"tags"`
		Watchers struct {
			Href string `json:"href"`
		} `json:"watchers"`
	} `json:"links"`
	Mainbranch struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"mainbranch"`
	Name  string `json:"name"`
	Owner struct {
		DisplayName string `json:"display_name"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Type     string `json:"type"`
		Username string `json:"username"`
		UUID     string `json:"uuid"`
	} `json:"owner"`
	Project struct {
		Key   string `json:"key"`
		Links struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Name string `json:"name"`
		Type string `json:"type"`
		UUID string `json:"uuid"`
	} `json:"project"`
	Scm       string    `json:"scm"`
	Size      int64     `json:"size"`
	Slug      string    `json:"slug"`
	Type      string    `json:"type"`
	UpdatedOn time.Time `json:"updated_on"`
	UUID      string    `json:"uuid"`
	Website   string    `json:"website"`
	Workspace struct {
		Links struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Type string `json:"type"`
		UUID string `json:"uuid"`
	} `json:"workspace"`
}

type userResponse struct {
	AccountID     string      `json:"account_id"`
	AccountStatus string      `json:"account_status"`
	CreatedOn     time.Time   `json:"created_on"`
	Department    interface{} `json:"department"`
	DisplayName   string      `json:"display_name"`
	Has2FaEnabled interface{} `json:"has_2fa_enabled"`
	IsStaff       bool        `json:"is_staff"`
	JobTitle      string      `json:"job_title"`
	Links         struct {
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
		Hooks struct {
			Href string `json:"href"`
		} `json:"hooks"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Repositories struct {
			Href string `json:"href"`
		} `json:"repositories"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Snippets struct {
			Href string `json:"href"`
		} `json:"snippets"`
	} `json:"links"`
	Location     interface{} `json:"location"`
	Nickname     string      `json:"nickname"`
	Organization interface{} `json:"organization"`
	Properties   struct {
	} `json:"properties"`
	Type     string      `json:"type"`
	UUID     string      `json:"uuid"`
	Zoneinfo interface{} `json:"zoneinfo"`
}

type prResponse struct {
	Author struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Nickname string `json:"nickname"`
		Type     string `json:"type"`
		UUID     string `json:"uuid"`
	} `json:"author"`
	CloseSourceBranch bool `json:"close_source_branch"`
	ClosedBy          struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Nickname string `json:"nickname"`
		Type     string `json:"type"`
		UUID     string `json:"uuid"`
	} `json:"closed_by"`
	CommentCount int64     `json:"comment_count"`
	CreatedOn    time.Time `json:"created_on"`
	Description  string    `json:"description"`
	Destination  struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
		Commit struct {
			Hash  string `json:"hash"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
			Type string `json:"type"`
		} `json:"commit"`
		Repository struct {
			FullName string `json:"full_name"`
			Links    struct {
				Avatar struct {
					Href string `json:"href"`
				} `json:"avatar"`
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
			Name string `json:"name"`
			Type string `json:"type"`
			UUID string `json:"uuid"`
		} `json:"repository"`
	} `json:"destination"`
	ID    int64 `json:"id"`
	Links struct {
		Activity struct {
			Href string `json:"href"`
		} `json:"activity"`
		Approve struct {
			Href string `json:"href"`
		} `json:"approve"`
		Comments struct {
			Href string `json:"href"`
		} `json:"comments"`
		Commits struct {
			Href string `json:"href"`
		} `json:"commits"`
		Decline struct {
			Href string `json:"href"`
		} `json:"decline"`
		Diff struct {
			Href string `json:"href"`
		} `json:"diff"`
		Diffstat struct {
			Href string `json:"href"`
		} `json:"diffstat"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Merge struct {
			Href string `json:"href"`
		} `json:"merge"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Statuses struct {
			Href string `json:"href"`
		} `json:"statuses"`
	} `json:"links"`
	MergeCommit struct {
		Hash string `json:"hash"`
	} `json:"merge_commit"`
	Participants []struct {
		Role           string    `json:"role"`
		Approved       bool      `json:"approved"`
		ParticipatedOn time.Time `json:"participated_on"`
		User           struct {
			AccountID string `json:"account_id"`
			UUID      string `json:"uuid"`
		} `json:"user"`
	} `json:"participants"`
	Reason string `json:"reason"`
	Source struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
		Commit struct {
			Hash  string `json:"hash"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
			Type string `json:"type"`
		} `json:"commit"`
		Repository struct {
			FullName string `json:"full_name"`
			Links    struct {
				Avatar struct {
					Href string `json:"href"`
				} `json:"avatar"`
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
			Name string `json:"name"`
			Type string `json:"type"`
			UUID string `json:"uuid"`
		} `json:"repository"`
	} `json:"source"`
	State   string `json:"state"`
	Summary struct {
		HTML   string `json:"html"`
		Markup string `json:"markup"`
		Raw    string `json:"raw"`
		Type   string `json:"type"`
	} `json:"summary"`
	TaskCount int64     `json:"task_count"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	UpdatedOn time.Time `json:"updated_on"`
}

type prCommentResponse struct {
	Content struct {
		HTML   string `json:"html"`
		Markup string `json:"markup"`
		Raw    string `json:"raw"`
		Type   string `json:"type"`
	} `json:"content"`
	CreatedOn time.Time `json:"created_on"`
	Deleted   bool      `json:"deleted"`
	ID        int64     `json:"id"`
	Links     struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	Pullrequest struct {
		ID    int64 `json:"id"`
		Links struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Title string `json:"title"`
		Type  string `json:"type"`
	} `json:"pullrequest"`
	Type      string    `json:"type"`
	UpdatedOn time.Time `json:"updated_on"`
	User      struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Nickname string `json:"nickname"`
		Type     string `json:"type"`
		UUID     string `json:"uuid"`
	} `json:"user"`
}
type prCommitResponse struct {
	Author struct {
		Raw  string `json:"raw"`
		Type string `json:"type"`
		User struct {
			AccountID   string `json:"account_id"`
			DisplayName string `json:"display_name"`
			Links       struct {
				Avatar struct {
					Href string `json:"href"`
				} `json:"avatar"`
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
			Nickname string `json:"nickname"`
			Type     string `json:"type"`
			UUID     string `json:"uuid"`
		} `json:"user"`
	} `json:"author"`
	Date  time.Time `json:"date"`
	Hash  string    `json:"hash"`
	Links struct {
		Approve struct {
			Href string `json:"href"`
		} `json:"approve"`
		Comments struct {
			Href string `json:"href"`
		} `json:"comments"`
		Diff struct {
			Href string `json:"href"`
		} `json:"diff"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Patch struct {
			Href string `json:"href"`
		} `json:"patch"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Statuses struct {
			Href string `json:"href"`
		} `json:"statuses"`
	} `json:"links"`
	Message string `json:"message"`
	Parents []struct {
		Hash  string `json:"hash"`
		Links struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Type string `json:"type"`
	} `json:"parents"`
	Repository struct {
		FullName string `json:"full_name"`
		Links    struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
		Name string `json:"name"`
		Type string `json:"type"`
		UUID string `json:"uuid"`
	} `json:"repository"`
	Summary struct {
		HTML   string `json:"html"`
		Markup string `json:"markup"`
		Raw    string `json:"raw"`
		Type   string `json:"type"`
	} `json:"summary"`
	Type string `json:"type"`
}

type webhookPayload struct {
	Active      bool     `json:"active"`
	CreatorID   string   `json:"creator_id"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	SubjectKey  string   `json:"subject_key"`
	URL         string   `json:"url"`
}
