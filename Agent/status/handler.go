package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Statuscheck(g *gin.Context) {
	var gpuUtil, gpumemUtil float32
	var err error
	gpuUtil, gpumemUtil, err = gpuinfo()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}
	g.JSON(http.StatusOK, gin.H{
		"gpuUtil":    gpuUtil,
		"gpumemUtil": gpumemUtil,
	})
}

func Probe(g *gin.Context) {
	// `dd if=/dev/zero of=probe_100k bs=102400 count=1`
	g.File("./probe_100k")
}
