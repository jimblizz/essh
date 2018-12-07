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

type ContainerDigest struct {
    Container string
    Status string
    Instance string
}

func (d *ContainerDigest) Load (t *ecs.Task, c *ecs.Container, instanceMap map[string]InstanceDigest) {
   d.Container = *c.Name
   d.Status = *c.LastStatus
   d.Instance = *t.ContainerInstanceArn

   if val, ok := instanceMap[d.Instance]; ok {
       d.Instance = val.Ec2InstanceId
   }
}

type InstanceDigest struct {
    ContainerInstanceArn string
    Ec2InstanceId string
}

func (d *InstanceDigest) Load (i *ecs.ContainerInstance) {
    d.ContainerInstanceArn = *i.ContainerInstanceArn
    d.Ec2InstanceId = *i.Ec2InstanceId

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

func GetClusterInstanceMap (ps PathStructure) (instancesMap map[string]InstanceDigest, err error) {

    instancesMap = make(map[string]InstanceDigest)

    // New ECS client
    svc, err := NewEcsClient(ps)
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    listInput := &ecs.ListContainerInstancesInput{
       Cluster: &ps.Cluster,
       MaxResults: aws.Int64(100),
    }
    list, err := svc.ListContainerInstances(listInput)
    if err != nil {
       return
    }

    describeInput := &ecs.DescribeContainerInstancesInput{
        Cluster: &ps.Cluster,
        ContainerInstances: list.ContainerInstanceArns,
    }
    result, err := svc.DescribeContainerInstances(describeInput)

    for _, ins := range result.ContainerInstances {
        var d InstanceDigest
        d.Load(ins)
        instancesMap[d.ContainerInstanceArn] = d
    }

    return
}

func GetContainerList (ps PathStructure, instanceMap map[string]InstanceDigest) (containers []ContainerDigest, err error) {

    // New ECS client
    svc, err := NewEcsClient(ps)
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    // TODO: Validate service name?

    // Get a list of tasks
    listInput := &ecs.ListTasksInput{
        Cluster: &ps.Cluster,
        ServiceName: &ps.Service,
        MaxResults: aws.Int64(100),
    }

    list, err := svc.ListTasks(listInput)
    if err != nil {
        return
    }

    // Get more data on these services
    describeInput := &ecs.DescribeTasksInput{
        Cluster: &ps.Cluster,
        Tasks: list.TaskArns,
    }
    result, err := svc.DescribeTasks(describeInput)

    // Build digests
    for _, t := range result.Tasks {
        for _, c := range t.Containers {
            var d ContainerDigest
            d.Load(t, c, instanceMap)
            containers = append(containers, d)
        }
    }

    return
}