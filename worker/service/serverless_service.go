package service

import (
	"context"
	"encoding/json"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"synapse/common"
	"synapse/worker/repository/repo"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
	"time"
)

type ServerlessService struct {
	db *gorm.DB
}

func NewServerlessService(db *gorm.DB) *ServerlessService {
	return &ServerlessService{db: db}
}

func (svc *ServerlessService) FindByEndpointId(ctx context.Context, endpointId string) (*types.ServerlessResourceResponse, error) {

	repository := repo.NewServerlessResourceRepository(svc.db.WithContext(ctx))

	result, exist, err := repository.FindByEndpointId(endpointId)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}

	res := new(types.ServerlessResourceResponse)
	err = copier.Copy(res, result)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (svc *ServerlessService) Create(ctx context.Context, req *types.CreateServerlessResourceRequest) (*types.ServerlessResourceResponse, error) {

	repository := repo.NewServerlessResourceRepository(svc.db.WithContext(ctx))

	record := &entity.ServerlessResource{
		EndpointId: req.EndpointId,
		Model:      req.Model,
		Status:     common.StatusSuccess,
		CreatedAt:  time.Now(),
	}

	result, err := repository.SaveOrUpdate(record)
	if err != nil {
		return nil, err
	}

	res := new(types.ServerlessResourceResponse)
	err = copier.Copy(res, result)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (svc *ServerlessService) Update(ctx context.Context, endpointId string, status int) (*types.ServerlessResourceResponse, error) {

	repository := repo.NewServerlessResourceRepository(svc.db.WithContext(ctx))

	record := &entity.ServerlessResource{
		EndpointId: endpointId,
		Status:     status,
		UpdatedAt:  time.Now(),
	}

	result, err := repository.SaveOrUpdate(record)
	if err != nil {
		return nil, err
	}

	res := new(types.ServerlessResourceResponse)
	err = copier.Copy(res, result)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Inference is a mock function for inference
func (svc *ServerlessService) Inference(ctx context.Context, endpointId string, req *types.InferenceMessageRequest) (*types.ChatCompletion, error) {

	repository := repo.NewServerlessResourceRepository(svc.db.WithContext(ctx))

	_, exist, err := repository.FindByEndpointId(endpointId)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, common.ErrEndpointNotFound
	}

	jsonString := `{
        "id": "chat-507887ec85af4eea8485fd257b4ba536",
        "object": "chat.completion",
        "created": 1733752587,
        "Model": "meta-llama/Llama-3.2-3B-Instruct",
        "choices": [
            {
                "index": 0,
                "message": {
                    "role": "assistant",
                    "content": "The Los Angeles Dodgers won the 2020 World Series. They defeated the Tampa Bay Rays in six games (4-2).",
                    "tool_calls": []
                },
                "logprobs": null,
                "finish_reason": "stop",
                "stop_reason": null
            }
        ],
        "usage": {
            "prompt_tokens": 51,
            "total_tokens": 78,
            "completion_tokens": 27
        },
        "prompt_logprobs": null
    }`

	var chatCompletion types.ChatCompletion
	err = json.Unmarshal([]byte(jsonString), &chatCompletion)
	if err != nil {
		return nil, err
	}

	return &chatCompletion, nil
}
