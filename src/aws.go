package main

import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws"
)

func CreateAwsSessionForProfileRegion (profile string, region string) (sess *session.Session, err error) {

    // Get based on profile selections?

    sess = session.Must(session.NewSessionWithOptions(session.Options{
        Config: aws.Config{Region: aws.String("eu-west-1")},
        Profile: "jim-tech",
    }))

    return
}

func CreateAwsSession () {

}
