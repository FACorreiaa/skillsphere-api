package example

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/FACorreiaa/skillsphere-proto/gen/myservice"
	"github.com/FACorreiaa/skillsphere-proto/gen/myservice/myserviceconnect"
)

// MyServiceHandler implements the Connect RPC handler for MyService
type MyServiceHandler struct {
	// Embed UnimplementedMyServiceHandler to ensure forward compatibility
	// If new RPCs are added to the proto, they'll return "unimplemented" by default
	myserviceconnect.UnimplementedMyServiceHandler

	service MyServiceService
	logger  *slog.Logger
}

// NewMyServiceHandler creates a new MyServiceHandler
func NewMyServiceHandler(svc MyServiceService, logger *slog.Logger) *MyServiceHandler {
	return &MyServiceHandler{
		service: svc,
		logger:  logger,
	}
}

// DoSomething handles the DoSomething RPC call
func (h *MyServiceHandler) DoSomething(
	ctx context.Context,
	req *connect.Request[myservice.MyServiceDoSomethingRequest],
) (*connect.Response[myservice.MyServiceDoSomethingResponse], error) {
	h.logger.Info("DoSomething called",
		"input", req.Msg.Input,
	)

	// Call the service layer
	result, err := h.service.DoSomething(ctx, req.Msg.Input)
	if err != nil {
		h.logger.Error("DoSomething failed",
			"error", err,
			"input", req.Msg.Input,
		)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Build response
	resp := &myservice.MyServiceDoSomethingResponse{
		Output: result,
	}

	h.logger.Info("DoSomething completed",
		"output", result,
	)

	return connect.NewResponse(resp), nil
}
