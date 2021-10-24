package status

import (
	"github.com/gin-gonic/gin"
)

func Probe(g *gin.Context) {
	// `dd if=/dev/zero of=probe_100k bs=102400 count=1`
	g.File("./probe_100k")
}
