package grpc

import (
	"context"

	"goload/internal/generated/grpc/goload"
	"goload/internal/logic"

	"google.golang.org/grpc"
)

type Handler struct {
	goload.UnimplementedGoLoadServiceServer
	accountService logic.AccountService
}

func NewHandler(
	accountService logic.AccountService,
) goload.GoLoadServiceServer {
	return &Handler{
		accountService: accountService,
	}
}

// CreateAccount implements goload.GoLoadServiceServer.
func (handler *Handler) CreateAccount(ctx context.Context, request *goload.CreateAccountRequest) (*goload.CreateAccountResponse, error) {
	output, err := handler.accountService.CreateAccount(ctx, logic.CreateAccountInput{
		AccountName: request.GetAccountName(),
		Password:    request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &goload.CreateAccountResponse{
		AccountId: output.ID,
	}, nil
}

// CreateDownloadTask implements goload.GoLoadServiceServer.
func (h *Handler) CreateDownloadTask(context.Context, *goload.CreateDownloadTaskRequest) (*goload.CreateDownloadTaskResponse, error) {
	panic("unimplemented")
}

// CreateSession implements goload.GoLoadServiceServer.
func (h *Handler) CreateSession(context.Context, *goload.CreateSessionRequest) (*goload.CreateSessionResponse, error) {
	panic("unimplemented")
}

// DeleteDownloadTask implements goload.GoLoadServiceServer.
func (h *Handler) DeleteDownloadTask(context.Context, *goload.DeleteDownloadTaskRequest) (*goload.DeleteDownloadTaskResponse, error) {
	panic("unimplemented")
}

// GetDownloadTaskFile implements goload.GoLoadServiceServer.
func (h *Handler) GetDownloadTaskFile(*goload.GetDownloadTaskFileRequest, grpc.ServerStreamingServer[goload.GetDownloadTaskFileResponse]) error {
	panic("unimplemented")
}

// GetDownloadTaskList implements goload.GoLoadServiceServer.
func (h *Handler) GetDownloadTaskList(context.Context, *goload.GetDownloadTaskListRequest) (*goload.GetDownloadTaskListResponse, error) {
	panic("unimplemented")
}

// UpdateDownloadTask implements goload.GoLoadServiceServer.
func (h *Handler) UpdateDownloadTask(context.Context, *goload.UpdateDownloadTaskRequest) (*goload.UpdateDownloadTaskResponse, error) {
	panic("unimplemented")
}
