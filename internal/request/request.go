package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-playground/validator/v10"
	"github.com/softball-manager/common/pkg/dynamo"
)

type CreateTeamRequest struct {
	Name    string   `json:"name" validate:"required"`
	Players []string `json:"players"`
}

var (
	teamIDPathParameterPrefix = "Team%23"
	validPidRegex             = fmt.Sprintf(`^%s[a-zA-Z0-9-]+$`, dynamo.TeamIDPrefix)
)

func ValidatePathParameters(request events.APIGatewayProxyRequest) (string, error) {
	switch len(request.PathParameters) {
	case 0:
		return "", nil
	case 1:
		if tid, found := request.PathParameters["tid"]; found {
			tid = strings.Replace(tid, teamIDPathParameterPrefix, dynamo.TeamIDPrefix, 1)
			validFormat := regexp.MustCompile(validPidRegex).MatchString(tid)
			if !validFormat {
				return "", errors.New("tid is not formatted correctly")
			}
			return tid, nil
		}
		return "", errors.New("invalid path parameters")
	default:
		return "", errors.New("too many path parameters provided")
	}
}

func ValidateCreateTeamRequest(requestBody string) (*CreateTeamRequest, error) {
	var validRequest CreateTeamRequest

	err := json.Unmarshal([]byte(requestBody), &validRequest)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(&validRequest); err != nil {
		return nil, err
	}

	return &validRequest, nil
}
