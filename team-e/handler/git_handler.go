package handler

import (
	"feather/service"
	"feather/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GitHandler struct {
	gitService service.GitService
}

func NewGitHandler(gs service.GitService) *GitHandler {
	return &GitHandler{gitService: gs}
}

func (h *GitHandler) CreateRepo(ctx *gin.Context) {
	var req *types.RepoFromTemplateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response(ctx, http.StatusUnprocessableEntity, err.Error())
	} else if res, err := h.gitService.CreateRepoBasedTemplate(req); err != nil {
		response(ctx, http.StatusInternalServerError, err.Error())
	} else {
		response(ctx, http.StatusOK, res)
	}
}
