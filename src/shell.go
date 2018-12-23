package main

import (
    "bytes"
    "fmt"
    "golang.org/x/crypto/ssh"
    "io/ioutil"
    "net"
    "os"
    "time"
)

const (
    SshTimeout = 3
)

func shellOpenSession (ps PathStructure) {

    // First, get the containers info
    instanceMap, err := GetClusterInstanceMap(ps)
    if err != nil {
        HandleAwsError(err)
        return
    }

    containers, err := GetContainerList(ps, instanceMap)
    if err != nil {
        HandleAwsError(err)
        return
    }

    for _, c := range containers {
        if c.UUID == ps.Container {

            // Start an SSH session
            // TODO: Hardcoded for testing, this will be different per users, and thus will need to be configured some place
            sshKeyLocation := "/var/application/browserinfrastructure/keys"

            pk, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s.pem", sshKeyLocation, c.Instance.KeyName))
            signer, err := ssh.ParsePrivateKey(pk)
            if err != nil {
                panic(err)
            }

            config := &ssh.ClientConfig{
                User: "ec2-user",
                Auth: []ssh.AuthMethod{
                    ssh.PublicKeys(signer),
                },
                HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
                    return nil
                },
                Timeout:time.Second * SshTimeout,
            }

            client, err := ssh.Dial("tcp", c.Instance.PublicIpAddress + ":22", config)
            if err != nil {
                panic("Failed to dial: " + err.Error())
            }

            // Each ClientConn can support multiple interactive sessions,
            // represented by a Session.
            session, err := client.NewSession()
            if err != nil {
                panic("Failed to create session: " + err.Error())
            }
            defer session.Close()

            // Once a Session is created, you can execute a single command on
            // the remote side using the Run method.
            var b bytes.Buffer
            session.Stdout = &b
            if err := session.Run("docker ps"); err != nil {
                panic("Failed to run: " + err.Error())
            }
            fmt.Println(b.String())


            os.Exit(0)
        }
    }

    fmt.Println(fmt.Sprintf("Cannot find container %s", ps.Container))
    return
}