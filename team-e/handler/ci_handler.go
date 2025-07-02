package handler

import (
	"feather/service"
	"feather/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CiHandler struct {
	argoSensorService service.ArgoSensorService
}

func NewCiHandler(as service.ArgoSensorService) *CiHandler {
	return &CiHandler{argoSensorService: as}
}

func (h *CiHandler) CreateArgoCi(ctx *gin.Context) {
	var req *types.JobBasedJavaRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response(ctx, http.StatusUnprocessableEntity, err.Error())
	} else if err := h.argoSensorService.CreateArgoSensor(req); err != nil {
		response(ctx, http.StatusInternalServerError, err.Error())
	} else {
		response(ctx, http.StatusOK, "Success")
	}
}
