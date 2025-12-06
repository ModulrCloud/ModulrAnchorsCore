package http_pack

import (
	"fmt"
	"strconv"

	"github.com/modulrcloud/modulr-anchors-core/globals"
	"github.com/modulrcloud/modulr-anchors-core/utils"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func createRouter() fasthttp.RequestHandler {

	r := router.New()

	// Default API routes
	r.GET("/block/{id}", GetBlockById)
	r.GET("/aggregated_finalization_proof/{blockId}", GetAggregatedFinalizationProof)

	// Route to request ARP (anchor rotation proof), then aggregated them and get AARP(Aggregated Anchor Rotation Proof)
	r.POST("/request_anchor_rotation_proof", RequestAnchorRotationProof)
	// Route to accept AARP, put to mempool and include to blocks
	r.POST("/accept_aggregated_anchor_rotation_proof", AcceptAggregatedAnchorRotationProofs)

	// Route to accept ALFP (Aggregated Leader Finalization Proof) from modulr-core logic, put to mempool and include to blocks
	r.POST("/accept_aggregated_leader_finalization_proof", AcceptAggregatedLeaderFinalizationProof)

	return r.Handler
}

func CreateHTTPServer() {

	serverAddr := globals.CONFIGURATION.Interface + ":" + strconv.Itoa(globals.CONFIGURATION.Port)

	utils.LogWithTime(fmt.Sprintf("Server is starting at http://%s ...✅", serverAddr), utils.CYAN_COLOR)

	if err := fasthttp.ListenAndServe(serverAddr, createRouter()); err != nil {
		utils.LogWithTime(fmt.Sprintf("Error in server: %s", err), utils.RED_COLOR)
	}
}
