// Copyright © 2016- 2024 Sesame Network Technology all right reserved

package define

import (
	"chatwiki/internal/app/chatwiki/llm/adaptor"

	"github.com/zhimaAi/go_tools/msql"
)

func GetAzureHandler(config msql.Params, _ string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:       `azure`,
			EndPoint:   config[`api_endpoint`],
			APIVersion: config[`api_version`],
			APIKey:     config[`api_key`],
			Model:      config[`deployment_name`],
		},
	}
	return handler, nil
}

func GetClaudeHandler(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:       `claude`,
			APIKey:     config[`api_key`],
			APIVersion: config[`api_version`],
			Model:      useModel,
		},
	}
	return handler, nil
}

func GetGeminiHandler(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:   `gemini`,
			APIKey: config[`api_key`],
			Model:  useModel,
		},
	}
	return handler, nil
}

func GetYiyanHandler(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:      `baidu`,
			APIKey:    config[`api_key`],
			SecretKey: config[`secret_key`],
			Model:     useModel,
		},
	}
	return handler, nil
}

func GetTongyiHandler(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:   `ali`,
			APIKey: config[`api_key`],
			Model:  useModel,
		},
	}
	return handler, nil
}

func GetBaaiHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:     `baai`,
			EndPoint: config[`api_endpoint`],
			APIKey:   config[`api_key`],
			Model:    useModel,
		},
	}
	return handler, nil
}

func GetCohereHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:     `cohere`,
			EndPoint: config[`api_endpoint`],
			APIKey:   config[`api_key`],
			Model:    useModel,
		},
	}
	return handler, nil
}

func GetDeepseekHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:   `deepseek`,
			APIKey: config[`api_key`],
			Model:  useModel,
		},
	}
	return handler, nil
}
func GetJinaHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:   `jina`,
			APIKey: config[`api_key`],
			Model:  useModel,
		},
	}
	return handler, nil
}

func GetLingYiWanWuHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:   `lingyiwanwu`,
			APIKey: config[`api_key`],
			Model:  useModel,
		},
	}
	return handler, nil
}

func GetMoonShotHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:   `moonshot`,
			APIKey: config[`api_key`],
			Model:  useModel,
		},
	}
	return handler, nil
}

func GetOpenAIHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:   `openai`,
			APIKey: config[`api_key`],
			Model:  useModel,
		},
	}
	return handler, nil
}

func GetSparkHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:      `spark`,
			APIKey:    config[`api_key`],
			SecretKey: config[`secret_key`],
			APPID:     config[`app_id`],
			Model:     useModel,
		},
	}
	return handler, nil
}

func GetHunyuanHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:      `hunyuan`,
			APIKey:    config[`api_key`],
			SecretKey: config[`secret_key`],
			Region:    config[`region`],
			Model:     useModel,
		},
	}
	return handler, nil
}

func GetDoubaoHandle(config msql.Params, useModel string) (*ModelCallHandler, error) {
	handler := &ModelCallHandler{
		Meta: adaptor.Meta{
			Corp:      `doubao`,
			APIKey:    config[`api_key`],
			SecretKey: config[`secret_key`],
			Region:    config[`region`],
			Model:     useModel,
		},
	}
	return handler, nil
}
