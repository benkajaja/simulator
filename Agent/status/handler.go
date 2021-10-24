package status

import (
	"github.com/gin-gonic/gin"
)

// Probe handler
//
// Send back 100KB file. The file is made by excuting following command
//  `dd if=/dev/zero of=probe_100k bs=102400 count=1`
func Probe(g *gin.Context) {
	g.File("./probe_100k")
}
