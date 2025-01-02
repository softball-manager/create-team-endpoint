package main

import (
	"context"
	"fmt"
	"net/http"
	"softball-manager/create-team-endpoint/internal/appconfig"
	"softball-manager/create-team-endpoint/internal/repository"
	"softball-manager/create-team-endpoint/internal/request"
	"softball-manager/create-team-endpoint/internal/response"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	commonCfg "github.com/softball-manager/common/pkg/appconfig"
	"github.com/softball-manager/common/pkg/awsconfig"
	"github.com/softball-manager/common/pkg/dynamo"
	"github.com/softball-manager/common/pkg/log"
	"go.uber.org/zap"
)

var teamEndpoint = "team-endpoint"
var dynamoClient *dynamodb.Client
var appCfg *appconfig.AppConfig
var repo *repository.Repository

func init() {
	env := commonCfg.GetEnvironment()

	logger := log.GetLoggerWithEnv(log.InfoLevel, env)
	logger.Sugar().Infof("initializing %s", teamEndpoint)

	cfg, err := awsconfig.GetAWSConfig(context.TODO(), env)
	if err != nil {
		logger.Sugar().Fatalf("Unable to load SDK config: %v", err)
	}

	appCfg = appconfig.NewAppConfig(env, cfg, logger)
	appCfg.ReadEnvVars()

	dynamoClient = dynamo.CreateClient(appCfg)
	repo = repository.NewRespository(context.TODO(), appCfg, dynamoClient)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := appCfg.SetLogger(log.GetLoggerWithEnv(log.InfoLevel, appCfg.Env))
	logger.Info("recieved event")

	tid, err := request.ValidatePathParameters(req)
	if err != nil {
		logger.Error("error validating path parameters", zap.Error(err))
		return response.CreateBadRequestResponse(), nil
	}

	switch req.HTTPMethod {
	case http.MethodPost:
		if tid == "" {
			return handleCreateTeam(ctx, req.Body)
		}
		return handleUpdateTeam(ctx, tid, req.Body)
	case http.MethodGet:
		return handleGetTeam(ctx, tid)
	default:
		return response.CreateBadRequestResponse(), nil
	}
}

func handleCreateTeam(ctx context.Context, requestBody string) (events.APIGatewayProxyResponse, error) {
	tid := fmt.Sprintf("%s%s", dynamo.TeamIDPrefix, uuid.New())
	appCfg.Logger = appCfg.Logger.With(zap.String(log.TeamIDLogKey, tid))
	logger := appCfg.GetLogger()

	validatedRequest, err := request.ValidateCreateTeamRequest(requestBody)
	if err != nil {
		logger.Error("error validating request", zap.Error(err))
		return response.CreateBadRequestResponse(), nil
	}

	err = repo.PutTeam(tid, validatedRequest.Name, validatedRequest.Players)
	if err != nil {
		logger.Error("error putting player into db", zap.Error(err))
		return response.CreateInternalServerErrorResponse(), nil
	}

	return response.CreateSuccessfulCreateTeamResponse(tid), nil
}

func handleUpdateTeam(ctx context.Context, tid string, requestBody string) (events.APIGatewayProxyResponse, error) {
	return response.CreateSuccesfulUpdateTeamResponse(), nil
}

func handleGetTeam(ctx context.Context, tid string) (events.APIGatewayProxyResponse, error) {
	appCfg.Logger = appCfg.Logger.With(zap.String(log.PlayerIDLogKey, tid))
	logger := appCfg.GetLogger()
	logger.Info("request validated")

	t, err := repo.GetTeam(tid)
	if err != nil {
		logger.Error("error getting player from db", zap.Error(err))
		return response.CreateInternalServerErrorResponse(), nil
	}

	if t.PK == "" {
		return response.CreateResourceNotFoundResponse(), nil
	}

	return response.CreateSuccessfulGetTeamResponse(t), nil
}

func main() {
	lambda.Start(handler)
}
