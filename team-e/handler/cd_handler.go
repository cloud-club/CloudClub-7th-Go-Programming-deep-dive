package handler

import (
	"feather/service"
	"feather/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CdHandler struct {
	argoCdService service.ArgoCdService
}

func NewCdHandler(as service.ArgoCdService) *CdHandler {
	return &CdHandler{argoCdService: as}
}

func (h *CdHandler) CreateArgoCd(ctx *gin.Context) {
	var req *types.CreateCdRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response(ctx, http.StatusUnprocessableEntity, err.Error())
	} else if err := h.argoCdService.CreateProjectManifestRepo(req); err != nil {
		response(ctx, http.StatusInternalServerError, err.Error())
	} else {
		response(ctx, http.StatusOK, "Success")
	}
}
