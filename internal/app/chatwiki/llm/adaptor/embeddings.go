// Copyright © 2016- 2024 Sesame Network Technology all right reserved

package adaptor

import (
	"chatwiki/internal/app/chatwiki/llm/api/ali"
	"chatwiki/internal/app/chatwiki/llm/api/azure"
	"chatwiki/internal/app/chatwiki/llm/api/baai"
	"chatwiki/internal/app/chatwiki/llm/api/baichuan"
	"chatwiki/internal/app/chatwiki/llm/api/baidu"
	"chatwiki/internal/app/chatwiki/llm/api/cohere"
	"chatwiki/internal/app/chatwiki/llm/api/gemini"
	"chatwiki/internal/app/chatwiki/llm/api/hunyuan"
	"chatwiki/internal/app/chatwiki/llm/api/jina"
	"chatwiki/internal/app/chatwiki/llm/api/openai"
	"chatwiki/internal/app/chatwiki/llm/api/volcenginev2"
	"chatwiki/internal/app/chatwiki/llm/api/voyage"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tencentHunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

type ZhimaEmbeddingRequest struct {
	Input string `json:"input"`
}

type ZhimaEmbeddingResponse struct {
	Result []float64 `json:"result"`
	//Usage  ZhimaEmbeddingUsage `json:"usage"`
}

func (a *Adaptor) CreateEmbeddings(req ZhimaEmbeddingRequest) (ZhimaEmbeddingResponse, error) {
	if req.Input == "" {
		return ZhimaEmbeddingResponse{}, errors.New("input empty")
	}

	switch a.meta.Corp {
	case "openai":
		client := openai.NewClient("https://api.openai.com/v1", a.meta.APIKey, &openai.ErrorResponse{})
		r := openai.EmbeddingRequest{
			Model: a.meta.Model,
			Input: []string{req.Input},
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	case "azure":
		client := azure.NewClient(
			a.meta.EndPoint,
			a.meta.APIVersion,
			a.meta.APIKey,
			a.meta.Model,
		)
		r := azure.EmbeddingRequest{
			Input: []string{req.Input},
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	case "baidu":
		client := baidu.NewClient(
			a.meta.APIKey,
			a.meta.SecretKey,
			a.meta.Model,
		)
		r := baidu.EmbeddingRequest{
			Input: []string{req.Input},
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	case "ali":
		client := ali.NewClient(a.meta.APIKey)
		r := ali.EmbeddingRequest{
			Input:      ali.Texts{Texts: []string{req.Input}},
			Model:      a.meta.Model,
			Parameters: ali.TextType{TextType: "document"},
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Output.Embeddings[0].Embedding,
		}, nil
	case "voyage":
		client := voyage.NewClient(
			a.meta.APIKey,
		)
		r := voyage.EmbeddingRequest{
			Input: []string{req.Input},
			Model: a.meta.Model,
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	case "gemini":
		client := gemini.NewClient(
			a.meta.APIKey,
			a.meta.Model,
		)
		r := gemini.EmbeddingRequest{
			Content: gemini.Content{Parts: []gemini.Part{{Text: req.Input}}},
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Embedding.Values,
		}, nil
	case "baichuan":
		client := baichuan.NewClient(a.meta.APIKey)
		r := openai.EmbeddingRequest{
			Model: a.meta.Model,
			Input: []string{req.Input},
		}
		res, err := client.OpenAIClient.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	case "baai":
		client := baai.NewClient(a.meta.EndPoint, a.meta.Model, a.meta.APIKey)
		r := baai.EmbeddingRequest{
			Model: a.meta.Model,
			Input: []string{req.Input},
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	case "volcengine":
		client := volcenginev2.NewClient(a.meta.EndPoint, a.meta.Model, a.meta.APIKey, a.meta.SecretKey, a.meta.Region)
		r := volcenginev2.EmbeddingRequest{
			Input: []string{req.Input},
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	case "cohere":
		client := cohere.NewClient(a.meta.APIKey)
		r := cohere.EmbeddingRequest{
			Texts:     []string{req.Input},
			Model:     a.meta.Model,
			InputType: "classification",
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Embeddings[0],
		}, nil
	case "tencent":
		client := hunyuan.NewClient(a.meta.APIKey, a.meta.SecretKey, a.meta.Region)
		r := tencentHunyuan.NewGetEmbeddingRequest()
		r.Input = common.StringPtr(req.Input)
		res, err := client.CreateEmbeddings(*r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		var result []float64
		for _, v := range res.Data[0].Embedding {
			result = append(result, *v)
		}
		return ZhimaEmbeddingResponse{
			Result: result,
		}, nil
	case "jina":
		client := jina.NewClient(a.meta.APIKey)
		r := jina.EmbeddingRequest{
			Input:        []string{req.Input},
			Model:        a.meta.Model,
			EncodingType: "float",
		}
		res, err := client.CreateEmbeddings(r)
		if err != nil {
			return ZhimaEmbeddingResponse{}, err
		}
		return ZhimaEmbeddingResponse{
			Result: res.Data[0].Embedding,
		}, nil
	}
	return ZhimaEmbeddingResponse{}, nil
}
