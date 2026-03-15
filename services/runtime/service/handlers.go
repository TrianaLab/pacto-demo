package service

import (
	"context"
)

type PredictInput struct {
	Body struct {
		Data []float64 `json:"data" doc:"Input data for prediction" minItems:"1"`
	}
}

type PredictOutput struct {
	Body struct {
		Prediction []float64 `json:"prediction" doc:"Model prediction results"`
		ModelID    string    `json:"model_id" doc:"Identifier of the model used"`
	}
}

func Predict(_ context.Context, input *PredictInput) (*PredictOutput, error) {
	resp := &PredictOutput{}
	resp.Body.Prediction = make([]float64, len(input.Body.Data))
	for i, v := range input.Body.Data {
		resp.Body.Prediction[i] = v * 2.0
	}
	resp.Body.ModelID = "demo-model-v1"
	return resp, nil
}

type ModelsOutput struct {
	Body struct {
		Models []ModelInfo `json:"models" doc:"List of available models"`
	}
}

type ModelInfo struct {
	ID      string `json:"id" doc:"Model identifier"`
	Name    string `json:"name" doc:"Model name"`
	Version string `json:"version" doc:"Model version"`
}

func ListModels(_ context.Context, _ *struct{}) (*ModelsOutput, error) {
	resp := &ModelsOutput{}
	resp.Body.Models = []ModelInfo{
		{ID: "demo-model-v1", Name: "Demo Model", Version: "1.0.0"},
	}
	return resp, nil
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status" doc:"Service health status"`
	}
}

func HealthCheck(_ context.Context, _ *struct{}) (*HealthOutput, error) {
	resp := &HealthOutput{}
	resp.Body.Status = "healthy"
	return resp, nil
}
