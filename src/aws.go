package main

import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ecs"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "fmt"
)

func NewAwsSession (ps PathStructure) (sess *session.Session, err error) {
    // Get based on profile selections?
    sess = session.Must(session.NewSessionWithOptions(session.Options{
        Config:  aws.Config{Region: aws.String(ps.Region)},
        Profile: ps.Profile,
    }))

    return
}

func NewEcsClient (ps PathStructure) (svc *ecs.ECS, err error) {
    sess, err := NewAwsSession(ps)
    if err != nil {
        return
    }
    svc = ecs.New(sess)
    return
}

func GetClusterList (ps PathStructure) (clusters *ecs.DescribeClustersOutput, err error) {
    // New ECS client
    svc, err := NewEcsClient(ps)
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    // Get a list of clusters
    listInput := &ecs.ListClustersInput{}
    list, err := svc.ListClusters(listInput)
    if err != nil {
        return
    }

    // Get more data on these clusters
    describeInput := &ecs.DescribeClustersInput{
        Clusters: list.ClusterArns,
    }
    clusters, err = svc.DescribeClusters(describeInput)
    return
}

func HandleAwsError (err error) {
    if aerr, ok := err.(awserr.Error); ok {
        switch aerr.Code() {
        case ecs.ErrCodeServerException:
            fmt.Println(ecs.ErrCodeServerException, aerr.Error())
        case ecs.ErrCodeClientException:
            fmt.Println(ecs.ErrCodeClientException, aerr.Error())
        case ecs.ErrCodeInvalidParameterException:
            fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
        default:
            fmt.Println(aerr.Error())
        }
    } else {
        // Print the error, cast err to awserr.Error to get the Code and
        // Message from an error.
        fmt.Println(err.Error())
    }
    return
}