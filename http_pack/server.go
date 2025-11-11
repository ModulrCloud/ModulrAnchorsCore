package http_pack

import (
	"fmt"
	"strconv"

	"github.com/ModulrCloud/ModulrAnchorsCore/globals"
	"github.com/ModulrCloud/ModulrAnchorsCore/routes"
	"github.com/ModulrCloud/ModulrAnchorsCore/utils"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func createRouter() fasthttp.RequestHandler {

	r := router.New()

	r.GET("/block/{id}", routes.GetBlockById)

	r.GET("/aggregated_finalization_proof/{blockId}", routes.GetAggregatedFinalizationProof)

	return r.Handler
}

func CreateHTTPServer() {

	serverAddr := globals.CONFIGURATION.Interface + ":" + strconv.Itoa(globals.CONFIGURATION.Port)

	utils.LogWithTime(fmt.Sprintf("Server is starting at http://%s ...✅", serverAddr), utils.CYAN_COLOR)

	if err := fasthttp.ListenAndServe(serverAddr, createRouter()); err != nil {
		utils.LogWithTime(fmt.Sprintf("Error in server: %s", err), utils.RED_COLOR)
	}
}
