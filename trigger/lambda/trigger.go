package lambda

import (
	"context"
	"encoding/json"
	"flag"

	"github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	syslog "log"
)

// log is the default package logger
var log = logger.GetLogger("trigger-tibco-lambda")
var singleton *LambdaTrigger

// LambdaTrigger AWS Lambda trigger struct
type LambdaTrigger struct {
	metadata *trigger.Metadata
	runner   action.Runner
	config   *trigger.Config
}

//NewFactory create a new Trigger factory
func NewFactory(md *trigger.Metadata) trigger.Factory {
	return &LambdaFactory{metadata: md}
}

// LambdaFactory AWS Lambda Trigger factory
type LambdaFactory struct {
	metadata *trigger.Metadata
}

//New Creates a new trigger instance for a given id
func (t *LambdaFactory) New(config *trigger.Config) trigger.Trigger {
	singleton = &LambdaTrigger{metadata: t.metadata, config: config}
	return singleton
}

// Metadata implements trigger.Trigger.Metadata
func (t *LambdaTrigger) Metadata() *trigger.Metadata {
	return t.metadata
}

func (t *LambdaTrigger) Init(runner action.Runner) {
	t.runner = runner
}

func Invoke() (interface{}, error) {

	log.Info("Starting AWS Lambda Trigger")
	// Use syslog since aws logs are still not that good
	syslog.Println("Starting AWS Lambda Trigger")

	// Parse the flags
	flag.Parse()

	// Looking up the arguments
	evtArg := flag.Lookup("evt")
	//evt := evtArg.Value.String()
	var evt interface{}
	// Unmarshall ctx
	if err := json.Unmarshal([]byte(evtArg.Value.String()), &evt); err != nil {
		return nil, err
	}

	log.Debugf("Received evt: '%+v'\n", evt)
	syslog.Printf("Received evt: '%+v'\n", evt)

	ctxArg := flag.Lookup("ctx")
	var lambdaCtx *runtime.Context

	// Unmarshal ctx
	if err := json.Unmarshal([]byte(ctxArg.Value.String()), &lambdaCtx); err != nil {
		return nil, err
	}

	log.Debugf("Received ctx: '%+v'\n", lambdaCtx)
	syslog.Printf("Received ctx: '%+v'\n", lambdaCtx)

	actionId := singleton.config.Handlers[0].ActionId
	log.Debugf("Calling actionid: '%s'\n", actionId)


	data := map[string]interface{}{
		"logStreamName":   lambdaCtx.LogStreamName,
		"logGroupName":    lambdaCtx.LogGroupName,
		"awsRequestId":    lambdaCtx.AWSRequestID,
		"memoryLimitInMB": lambdaCtx.MemoryLimitInMB,
		"evt":             evt,
	}

	startAttrs, err := singleton.metadata.OutputsToAttrs(data, false)
	if err != nil {
		log.Errorf("After run error' %s'\n", err)
		return nil, err
	}

	act := action.Get(actionId)
	ctx := trigger.NewContextWithData(context.Background(), &trigger.ContextData{Attrs: startAttrs, HandlerCfg: singleton.config.Handlers[0]})
	_, replyData, err := singleton.runner.Run(ctx, act, actionId, nil)

	if err != nil {
		log.Debugf("Lambda Trigger Error: %s", err.Error())
		return nil, err
	}

	return replyData, err
}

func (t *LambdaTrigger) Start() error {
	return nil
}

// Stop implements util.Managed.Stop
func (t *LambdaTrigger) Stop() error {
	return nil
}
