package main

import (
    "fmt"
    "github.com/tidwall/gjson"
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
            sshSession(c)
            return
        }
    }

    fmt.Println(fmt.Sprintf("Cannot find container %s", ps.Container))
    return
}

func sshSession (c ContainerDigest) {

    //spew.Dump(c)

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
        client.Close()
        panic("Failed to create session: " + err.Error())
    }
    defer session.Close()

    // TODO: Multiple commands: https://stackoverflow.com/questions/24440193/golang-ssh-how-to-run-multiple-commands-on-the-same-session
    //shellExec(session, "docker ps -q")
    //shellExec(session, "docker ps -a")
    containerId := shellGetDockerId(session, c)
    fmt.Println(containerId)

    // TODO: We need to be using an interactive SSH session to proceed

    os.Exit(0)
}

func shellGetDockerId (session *ssh.Session, c ContainerDigest) string {
    curlCommand := fmt.Sprintf("curl http://localhost:51678/v1/tasks?arn=%s", c.TaskArn)
    out, err := session.CombinedOutput(curlCommand)
    if err != nil {
        fmt.Println(err)
    }

    tasks := gjson.Get(string(out), "Tasks")

    if tasks.IsArray() {
        for _, task := range tasks.Array() {
            if task.Get("Arn").String() == c.TaskArn {

                containers := task.Get("Containers")
                if containers.IsArray() {
                    for _, container := range containers.Array() {

                        // In theory we should have unique names within a give tasks
                        // Other deploying versions would show as a different task, so we know this would be the correct container
                        if container.Get("Name").String() == c.Container {
                            containerId := container.Get("DockerId").String()
                            return containerId
                        }

                    }
                }

            }
        }
    }

    return ""
}