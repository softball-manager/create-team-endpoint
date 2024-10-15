package repository

import (
	"context"
	"softball-manager/create-team-endpoint/internal/pkg/appconfig"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/softball-manager/common/pkg/log"
	"github.com/softball-manager/common/pkg/team"
	"go.uber.org/zap"
)

type Repository struct {
	Ctx       context.Context
	AppConfig *appconfig.AppConfig
	Client    *dynamodb.Client
	TableName string
}

func NewRespository(ctx context.Context, cfg *appconfig.AppConfig, client *dynamodb.Client) *Repository {
	return &Repository{
		Ctx:       ctx,
		AppConfig: cfg,
		Client:    client,
		TableName: cfg.TableName,
	}
}

func (r *Repository) CreateTeam(pk string, name string, players []string) error {
	logger := r.AppConfig.GetLogger().With(zap.String(log.TableNameLogKey, r.TableName))
	t := team.Team{
		PK:       pk,
		SK:       pk,
		TeamName: name,
		Players:  players,
	}

	logger.Info("marshalling team struct")
	av, err := attributevalue.MarshalMap(t)
	if err != nil {
		return err
	}

	logger.Info("inserting item into db", zap.Any("item", av))
	_, err = r.Client.PutItem(r.Ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName),
		Item:      av,
	})
	if err != nil {
		return err
	}
	logger.Info("successfully inserted item")

	return nil
}
