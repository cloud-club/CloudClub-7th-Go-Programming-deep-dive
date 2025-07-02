package types

import (
	"strings"
	"time"
)

type header struct {
	Result int    `json:"result"`
	Data   string `json:"data"`
}

func newHeader(result int, data ...string) *header {
	return &header{
		Result: result,
		Data:   strings.Join(data, ", "),
	}
}

type Response struct {
	*header
	Result interface{} `json:"result"`
}

func NewRes(result int, res interface{}, data ...string) *Response {
	return &Response{
		header: newHeader(result, data...),
		Result: res,
	}
}

type GiteaFileResponse struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

type GiteaFileContentResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	URL      string `json:"url"`
	SHA      string `json:"sha"`
	Size     int64  `json:"size"`
}

type GiteaUser struct {
	Active            bool      `json:"active"`
	AvatarURL         string    `json:"avatar_url"`
	Created           time.Time `json:"created"`
	Description       string    `json:"description"`
	Email             string    `json:"email"`
	FollowersCount    int64     `json:"followers_count"`
	FollowingCount    int64     `json:"following_count"`
	FullName          string    `json:"full_name"`
	HTMLURL           string    `json:"html_url"`
	ID                int64     `json:"id"`
	IsAdmin           bool      `json:"is_admin"`
	Language          string    `json:"language"`
	LastLogin         time.Time `json:"last_login"`
	Location          string    `json:"location"`
	Login             string    `json:"login"`
	LoginName         string    `json:"login_name"`
	ProhibitLogin     bool      `json:"prohibit_login"`
	Restricted        bool      `json:"restricted"`
	SourceID          int64     `json:"source_id"`
	StarredReposCount int64     `json:"starred_repos_count"`
	Visibility        string    `json:"visibility"`
	Website           string    `json:"website"`
}
