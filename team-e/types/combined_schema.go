package types

type ProjectWithBaseCampInfo struct {
	BaseCampName  string `json:"baseCamp_name"`
	BaseCampURL   string `json:"baseCamp_url"`
	BaseCampOwner string `json:"baseCamp_owner"`
	Token         string `json:"token"`

	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name" binding:"required"`
	ProjectURL   string `json:"project_url" binding:"required"`
	ProjectOwner string `json:"project_owner" binding:"required"`
}
