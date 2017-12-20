package awsenv

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
	"github.com/kataras/golog"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
)

type SsmParameter struct {
	paramType string
	version	  string
	length    string
	value     string
}

func CreateSSMParameters(paramsFile string, region string) {

    params, err := loadParams(paramsFile)
    if err != nil {
		golog.Fatal(err)
	}

	ssmClient := createSSMClient(region)

	m := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(params), m)
	var paramMap map[string]string
	paramMap = flatmap.Flatten(m)
	for key, value := range paramMap {
		paramName := strings.Replace(key, ".", "/", -1)
		if strings.Contains(paramName, "/") {
			paramName = "/" + paramName
		}
		createParam(ssmClient, paramName, value)
	}
}

func loadParams(params string) ([]byte, error) {
	golog.Infof("loading ssm params for %s", params)
	return ioutil.ReadFile(params)
}

func createSSMClient(region string) *ssm.SSM {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ssm.New(sess)
	return svc
}

func createParam(ssmClient *ssm.SSM, paramName string, paramValue string) {
	if isParamNewOrUpdated(ssmClient, paramName, paramValue) {
		golog.Infof("creating or updating ssm parameter %s=%s", paramName, paramValue)
		input := &ssm.PutParameterInput{
			Name:      aws.String(paramName),
			Type:      aws.String("String"),
			Value:     aws.String(paramValue),
			Overwrite: aws.Bool(true),
		}
		_, err := ssmClient.PutParameter(input)
		if err != nil {
			golog.Fatalf("failed to create param %s", err)
		}
	} else {
		golog.Infof("ssm parameter %s already exists", paramName)
	}
}

func isParamNewOrUpdated(ssmClient *ssm.SSM, paramName string, paramValue string) bool {
	param := getParam(ssmClient, paramName)
	return param == nil || aws.StringValue(param.Value) != paramValue
}

func getParam(ssmClient *ssm.SSM, paramName string) *ssm.Parameter {
	result, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(paramName),
	})
	if err != nil {
		if strings.Contains(err.Error(), ssm.ErrCodeParameterNotFound) {
			return nil
		}
		golog.Fatalf("failed to get param %s caused by %s", paramName, err)
	}
	return result.Parameter
}

