package llm

import (
	"context"
	"fmt"
	"salesforce-ai-agent/configuration"
	"sync"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

func InitializeLLM(ctx context.Context, config *configuration.Config, appLogger *zap.Logger) (*LLMModels, error) {
	llmModels := LLMModels{
		Models: make(map[string]*LLMModel),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, modelIter := range config.Models.Models {
		wg.Add(1)
		go func(modelIter configuration.Model) {
			defer wg.Done()

			config := openai.DefaultConfig(modelIter.APIKey)
			config.BaseURL = modelIter.BaseURL
			client := openai.NewClientWithConfig(config)

			model := &LLMModel{
				Model:  modelIter.Model,
				Client: client,
			}

			mu.Lock()
			llmModels.Models[modelIter.Name] = model
			mu.Unlock()

		}(modelIter)
	}

	wg.Wait()

	if len(llmModels.Models) == 0 {
		return nil, fmt.Errorf("no models available — check your config and API keys")
	}

	return &llmModels, nil
}

func SetPlannerExecutor(models *LLMModels, configYaml *configuration.Config, appLogger *zap.Logger) (planner *LLMModel, executor *LLMModel, err error) {
	if len(models.Models) == 0 {
		return nil, nil, fmt.Errorf("no models available — check your config and API keys")
	}

	if len(models.Models) == 1 {
		for _, model := range models.Models {
			return model, model, nil
		}
	}

	if configYaml.Planner != "" {
		if models.Models[configYaml.Planner] != nil {
			planner = models.Models[configYaml.Planner]
		} else {
			appLogger.Error("Planner model not set",
				zap.Error(err),
			)
		}
	}

	if configYaml.Executor != "" {
		if models.Models[configYaml.Executor] != nil {
			executor = models.Models[configYaml.Executor]
		} else {
			appLogger.Error("Executor model not set",
				zap.Error(err),
			)
		}
	}

	if planner != nil && executor != nil {
		return planner, executor, nil
	}

	count := 0
	randomModels := make([]*LLMModel, 0)
	for _, model := range models.Models {
		if count == 2 {
			return randomModels[0], randomModels[1], nil
		}
		randomModels = append(randomModels, model)
		count++
	}

	return planner, executor, nil
}
