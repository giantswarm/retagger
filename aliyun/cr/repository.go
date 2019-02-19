package cr

type RepositoryInput struct {
	RepoNamespace string `json:"RepoNamespace"`
	RepoName      string `json:"RepoName"`
	Summary       string `json:"Summary"`
	Detail        string `json:"Detail"`
	RepoType      string `json:"RepoType"`
}

type RepositoryRequest struct {
	Repo RepositoryInput `json:"Repo"`
}

type Repository struct {
	Detail         string `json:"detail"`
	Summary        string `json:"summary"`
	Logo           string `json:"logo"`
	Stars          int    `json:"stars"`
	RepoDomainList struct {
		Internal string `json:"internal"`
		Public   string `json:"public"`
		Vpc      string `json:"vpc"`
	} `json:"repoDomainList"`
	RepoAuthorizeType string `json:"repoAuthorizeType"`
	Downloads         int    `json:"downloads"`
	RegionID          string `json:"regionId"`
	RepoType          string `json:"repoType"`
	RepoNamespace     string `json:"repoNamespace"`
	RepoName          string `json:"repoName"`
	RepoID            int    `json:"repoId"`
	RepoStatus        string `json:"repoStatus"`
	RepoOriginType    string `json:"repoOriginType"`
	GmtCreate         int64  `json:"gmtCreate"`
	RepoBuildType     string `json:"repoBuildType"`
	GmtModified       int64  `json:"gmtModified"`
}

type RepoResponse struct {
	Data struct {
		Repo Repository `json:"repo"`
	} `json:"data"`
	RequestID string `json:"requestId"`
}
