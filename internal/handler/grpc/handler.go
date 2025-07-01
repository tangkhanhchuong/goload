package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"goload/internal/generated/grpc/goload"
	"goload/internal/logic"
)

const (
	AuthTokenMetadataName = "goload-auth"
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
	account, err := handler.accountService.CreateAccount(ctx, logic.CreateAccountInput{
		AccountName: request.GetAccountName(),
		Password:    request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &goload.CreateAccountResponse{
		AccountId: account.ID,
	}, nil
}

// CreateSession implements goload.GoLoadServiceServer.
func (h *Handler) CreateSession(ctx context.Context, request *goload.CreateSessionRequest) (*goload.CreateSessionResponse, error) {
	session, err := h.accountService.CreateSession(ctx, logic.CreateSessionInput{
		AccountName: request.GetAccountName(),
		Password:    request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	err = grpc.SetHeader(ctx, metadata.Pairs(AuthTokenMetadataName, session.Token))
	if err != nil {
		return nil, err
	}

	return &goload.CreateSessionResponse{
		Account: &goload.Account{
			Id:          session.Account.Id,
			AccountName: session.Account.AccountName},
	}, nil
}

// CreateDownloadTask implements goload.GoLoadServiceServer.
func (h *Handler) CreateDownloadTask(context.Context, *goload.CreateDownloadTaskRequest) (*goload.CreateDownloadTaskResponse, error) {
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
