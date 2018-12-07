package main

import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ecs"
    "fmt"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "sort"
    "strings"
)


type ServiceDigest struct {
    ServiceName string
    DesiredCount int64
    RunningCount int64
    PendingCount int64
    LatestEvent string
    Status string
    Role string
    TaskDefinition string
}

func (d *ServiceDigest) Load (s ecs.Service) {
    d.ServiceName = *s.ServiceName
    d.RunningCount = *s.RunningCount
    d.DesiredCount = *s.DesiredCount
    d.PendingCount = *s.PendingCount
    d.LatestEvent = *s.Events[0].Message
    d.Status = *s.Status

    if s.RoleArn != nil {
        d.Role = *s.RoleArn
        i := strings.Index(d.Role, "/")
        d.Role = d.Role[i+1:]
    }

    if s.TaskDefinition != nil {
        d.TaskDefinition = *s.TaskDefinition
        i := strings.Index(d.TaskDefinition, "/")
        d.TaskDefinition = d.TaskDefinition[i+1:]
    }
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

    sort.Slice(clusters.Clusters[:], func(i, j int) bool {
        return *clusters.Clusters[i].ClusterName < *clusters.Clusters[j].ClusterName
    })

    return
}

func GetServiceList (ps PathStructure) (digest []ServiceDigest, err error) {

    // New ECS client
    svc, err := NewEcsClient(ps)
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    // Get a list of services
    // TODO: We are limited to 100 services per cluster, this could be an issue later?
    listInput := &ecs.ListServicesInput{
       Cluster: &ps.Cluster,
       MaxResults: aws.Int64(100),
    }
    list, err := svc.ListServices(listInput)
    if err != nil {
       return
    }

    var pageHolder []*string
    var chunkSize = 10
    var counter int

    for i, arn := range list.ServiceArns {
        pageHolder = append(pageHolder, arn)
        counter++

        if counter >= chunkSize || i >= len(list.ServiceArns)-1 {

            // Get more data on these services
            describeInput := &ecs.DescribeServicesInput{
               Cluster: &ps.Cluster,
               Services: pageHolder,
            }
            services, err := svc.DescribeServices(describeInput)
            if err != nil {
               HandleAwsError(err)
               break
            }

            for _, s := range services.Services {
               var d ServiceDigest
               d.Load(*s)
               digest = append(digest, d)
            }

            pageHolder = make([]*string, 0)
            counter = 0
        }
    }

    sort.Slice(digest[:], func(i, j int) bool {
        return digest[i].ServiceName < digest[j].ServiceName
    })

    return
}