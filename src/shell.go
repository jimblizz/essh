package main

import (
    "bufio"
    "fmt"
    "github.com/tidwall/gjson"
    "golang.org/x/crypto/ssh"
    "io"
    "io/ioutil"
    "net"
    "os"
    "os/signal"
    "sync"
    "syscall"
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
    //containerId := shellGetDockerId(session, c)
    //fmt.Println(containerId)

    // We need to be using an interactive SSH session to proceed
    modes := ssh.TerminalModes{
        ssh.ECHO:          0,     // disable echoing
        ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
        ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
    }

    if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
        panic(err)
    }

    w, err := session.StdinPipe()
    if err != nil {
        panic(err)
    }
    r, err := session.StdoutPipe()
    if err != nil {
        panic(err)
    }
    in, out := MuxShell(w, r)
    if err := session.Start("/bin/sh"); err != nil {
        panic(err)
    }

    <-out // Ignore the first shell output, this is from setting up the session

    // Listen for Ctrl-C events, and disconnect cleanly
    // Listen for sigterm
    // Watch for kill services
    var gracefulStop = make(chan os.Signal)
    signal.Notify(gracefulStop, syscall.SIGTERM)
    signal.Notify(gracefulStop, syscall.SIGINT)

    go func() {
        sig := <-gracefulStop
        fmt.Printf("Caught sig: %+v \n", sig)
        fmt.Println("Closing connection")

        session.Close()
        client.Close()
        os.Exit(0)
    }()

    // First we need to use the ECS agent introspection too to find the container ID that we require
    // We will do this with a cURL call from inside the host
    in <- fmt.Sprintf("curl http://localhost:51678/v1/tasks?arn=%s", c.TaskArn)
    agentData := <-out

    containerId := shellGetDockerId(string(agentData), c)
    if containerId == "" {
        fmt.Println("Could not extract DockerId for the container")
        session.Close()
        os.Exit(0)
    }

    in <- fmt.Sprintf("docker exec -it %s /bin/sh", containerId)
    fmt.Print(<-out)

    fmt.Printf("Connected to %s.\n", c.Container)
    fmt.Println("Type 'exit' to quit.")

    in <- "pwd"
    fmt.Print(<-out)

    for {
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("> ")
        text, _ := reader.ReadString('\n')

        if text == "exit\n" || text == "quit\n" {
            fmt.Println("Closing connection")
            break
        }

        in <- text
        fmt.Println(<-out)
    }

    in <- "exit" // Exit container
    fmt.Print(<-out)

    in <- "exit" // Exit host
    fmt.Print(<-out)

    session.Wait()

    os.Exit(0)
}

func shellGetDockerId (json string, c ContainerDigest) string {
    tasks := gjson.Get(string(json), "Tasks")

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

func MuxShell(w io.Writer, r io.Reader) (chan<- string, <-chan string) {
    in := make(chan string, 1)
    out := make(chan string, 1)
    var wg sync.WaitGroup
    wg.Add(1) // for the shell itself
    go func() {
        for cmd := range in {
            wg.Add(1)
            w.Write([]byte(cmd + "\n"))
            wg.Wait()
        }
    }()
    go func() {
        var (
            buf [65 * 1024]byte
            t   int
        )
        for {
            n, err := r.Read(buf[t:])
            if err != nil {
                close(in)
                close(out)
                return
            }
            t += n
            if buf[t-2] == '$' || buf[t-2] == '#' { // assuming the $PS1 == 'sh-4.3$ '
                //fmt.Printf("DEBUG: |%s|\n", string(buf[t-2:]))
                out <- string(buf[:t])
                t = 0
                wg.Done()
            }
        }
    }()
    return in, out
}