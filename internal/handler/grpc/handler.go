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
	accountService      logic.AccountService
	downloadTaskService logic.DownloadTaskService
	tokenService        logic.TokenService
}

func NewHandler(
	accountService logic.AccountService,
	downloadTaskService logic.DownloadTaskService,
	tokenService logic.TokenService,
) goload.GoLoadServiceServer {
	return &Handler{
		accountService:      accountService,
		downloadTaskService: downloadTaskService,
		tokenService:        tokenService,
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
	output, err := h.accountService.CreateSession(ctx, logic.CreateSessionInput{
		AccountName: request.GetAccountName(),
		Password:    request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &goload.CreateSessionResponse{
		Account: &goload.Account{
			Id:          output.Account.Id,
			AccountName: output.Account.AccountName},
		Token: output.Token,
	}, nil
}

// CreateDownloadTask implements goload.GoLoadServiceServer.
func (h *Handler) CreateDownloadTask(ctx context.Context, request *goload.CreateDownloadTaskRequest) (*goload.CreateDownloadTaskResponse, error) {
	accountID, _, err := h.tokenService.ParseAccountIDAndExpireTime(ctx, h.getAuthTokenMetadata(ctx))
	if err != nil {
		return nil, err
	}

	output, err := h.downloadTaskService.CreateDownloadTask(ctx, logic.CreateDownloadTaskInput{
		OfAccountID: accountID,
		URL:         request.GetUrl(),
	})
	if err != nil {
		return nil, err
	}

	return &goload.CreateDownloadTaskResponse{
		DownloadTask: output.DownloadTask,
	}, nil
}

// DeleteDownloadTask implements goload.GoLoadServiceServer.
func (h *Handler) DeleteDownloadTask(ctx context.Context, request *goload.DeleteDownloadTaskRequest) (*goload.DeleteDownloadTaskResponse, error) {
	accountID, _, err := h.tokenService.ParseAccountIDAndExpireTime(ctx, h.getAuthTokenMetadata(ctx))
	if err != nil {
		return nil, err
	}

	output, err := h.downloadTaskService.DeleteDownloadTask(ctx, logic.DeleteDownloadTaskInput{
		OfAccountID:    accountID,
		DownloadTaskID: request.GetId(),
	})
	if err != nil {
		return nil, err
	}

	return &goload.DeleteDownloadTaskResponse{
		Deleted: output.Deleted,
	}, nil
}

// GetDownloadTaskFile implements goload.GoLoadServiceServer.
func (h *Handler) GetDownloadTaskFile(*goload.GetDownloadTaskFileRequest, grpc.ServerStreamingServer[goload.GetDownloadTaskFileResponse]) error {
	panic("unimplemented")
}

// GetDownloadTaskList implements goload.GoLoadServiceServer.
func (h *Handler) GetDownloadTaskList(ctx context.Context, request *goload.GetDownloadTaskListRequest) (*goload.GetDownloadTaskListResponse, error) {
	accountID, _, err := h.tokenService.ParseAccountIDAndExpireTime(ctx, h.getAuthTokenMetadata(ctx))
	if err != nil {
		return nil, err
	}

	output, err := h.downloadTaskService.GetDownloadTaskList(ctx, logic.GetDownloadTaskListInput{
		OfAccountID: accountID,
		Offset:      request.GetOffset(),
		Limit:       request.GetLimit(),
	})
	if err != nil {
		return nil, err
	}

	return &goload.GetDownloadTaskListResponse{
		DownloadTaskList:       output.DownloadTaskList,
		TotalDownloadTaskCount: output.TotalDownloadTaskCount,
	}, nil
}

// UpdateDownloadTask implements goload.GoLoadServiceServer.
func (h *Handler) UpdateDownloadTask(ctx context.Context, request *goload.UpdateDownloadTaskRequest) (*goload.UpdateDownloadTaskResponse, error) {
	accountID, _, err := h.tokenService.ParseAccountIDAndExpireTime(ctx, h.getAuthTokenMetadata(ctx))
	if err != nil {
		return nil, err
	}

	output, err := h.downloadTaskService.UpdateDownloadTask(ctx, logic.UpdateDownloadTaskInput{
		DownloadTaskID:     request.GetId(),
		OfAccountID:        accountID,
		URL:                request.GetUrl(),
		DownloadTaskStatus: request.GetDownloadTaskStatus(),
	})
	if err != nil {
		return nil, err
	}

	return &goload.UpdateDownloadTaskResponse{
		Updated: output.Updated,
	}, nil
}

func (a Handler) getAuthTokenMetadata(ctx context.Context) string {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	metadataValues := metadata.Get(AuthTokenMetadataName)
	if len(metadataValues) == 0 {
		return ""
	}

	return metadataValues[0]
}
